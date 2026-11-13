package initialize

import (
	"reflect"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// softDeleteClauseKey 标记「本语句已追加过软删除条件」，避免重复回调叠加条件。
const softDeleteClauseKey = "ape:soft_delete"

// softDeleteTransformKey 标记「本删除语句已改写为逻辑删除」，避免回调重复执行。
const softDeleteTransformKey = "ape:soft_delete:transform"

// RegisterSoftDeleteFilter 让内嵌 model.SoftDeleted 的实体默认只命中未删除数据，
// 并把对它们的物理删除自动改写为逻辑删除。
//
// 查询和关联加载统一过滤 is_deleted；删除操作改写为 UPDATE，并填充删除审计字段。
// 使用 Unscoped() 可查询已删除数据或执行物理删除。
func RegisterSoftDeleteFilter(db *gorm.DB) {
	_ = db.Callback().Query().Before("gorm:query").Register("ape:soft_delete:query", applySoftDeleteFilter)
	_ = db.Callback().Update().Before("gorm:update").Register("ape:soft_delete:update", applySoftDeleteFilter)
	_ = db.Callback().Delete().Before("gorm:delete").Register("ape:soft_delete:delete", softDeleteOnDelete)
}

// applySoftDeleteFilter 给查询与更新追加 is_deleted = false。
//
// 更新也要加：否则对被软删除的记录执行 Update 仍会成功，删除语义形同虚设。
func applySoftDeleteFilter(tx *gorm.DB) {
	if tx.Error != nil || tx.Statement == nil || tx.Statement.Unscoped {
		return
	}
	field := softDeleteField(tx)
	if field == nil {
		return
	}
	if _, ok := tx.Statement.Clauses[softDeleteClauseKey]; ok {
		return
	}

	// 条件里已有 OR 时必须整体加括号，否则 `is_deleted = 0 AND a OR b`
	// 会退化成 `(is_deleted = 0 AND a) OR b`，把已删除数据带出来。
	if c, ok := tx.Statement.Clauses["WHERE"]; ok {
		if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) >= 1 {
			for _, expr := range where.Exprs {
				if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
					where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
					c.Expression = where
					tx.Statement.Clauses["WHERE"] = c
					break
				}
			}
		}
	}

	tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{
			Column: clause.Column{Table: clause.CurrentTable, Name: field.DBName},
			Value:  false,
		},
	}})
	tx.Statement.Clauses[softDeleteClauseKey] = clause.Clause{}
}

// softDeleteOnDelete 把对内嵌 SoftDeleted 实体的物理删除改写为逻辑删除。
//
// 这里复用 GORM 内建软删除（gorm.DeletedAt）的做法：先补主键条件，再只命中未删除行，
// 然后用 UPDATE 的 clause 构建器生成语句并写入 Statement.SQL。gorm:delete 回调
// 发现 SQL 已生成就会跳过 DELETE 的构建，直接执行这条 UPDATE，方言差异（引号、
// 布尔字面量）仍由 GORM 处理。
//
// 未内嵌 SoftDeleted 的实体（关联表等）保持物理删除；`Unscoped().Delete` 同样物理删除。
func softDeleteOnDelete(tx *gorm.DB) {
	if tx.Error != nil || tx.Statement == nil || tx.Statement.Unscoped {
		return
	}
	field := softDeleteField(tx)
	if field == nil {
		return
	}
	stmt := tx.Statement
	if _, ok := stmt.Clauses[softDeleteTransformKey]; ok {
		return
	}

	// `Delete(&Model{}, ids)` 的 id 已由 BuildCondition 写入 WHERE；
	// `Delete(&entity)` 则依赖 delete builder 从 ReflectValue 提取主键。
	// 这里改走 update builder，需要自己把主键条件补上。
	hasCondition := false
	if _, ok := stmt.Clauses["WHERE"]; ok {
		hasCondition = true
	}
	if stmt.Schema != nil {
		_, queryValues := schema.GetIdentityFieldValuesMap(stmt.Context, stmt.ReflectValue, stmt.Schema.PrimaryFields)
		column, values := schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)
		if len(values) > 0 {
			stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
			hasCondition = true
		}
		if stmt.ReflectValue.CanAddr() && stmt.Dest != stmt.Model && stmt.Model != nil {
			_, queryValues = schema.GetIdentityFieldValuesMap(stmt.Context, reflect.ValueOf(stmt.Model), stmt.Schema.PrimaryFields)
			column, values = schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)
			if len(values) > 0 {
				stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
				hasCondition = true
			}
		}
	}
	// 没有任何条件时拒绝执行，避免误删全表。
	if !hasCondition {
		tx.AddError(gorm.ErrMissingWhereClause)
		return
	}

	// 只处理未删除的行，重复删除不会刷新删除人/删除时间。
	applySoftDeleteFilter(tx)

	now := ext.GetCurrentTime()
	sets := clause.Set{
		{Column: clause.Column{Name: field.DBName}, Value: true},
		{Column: clause.Column{Name: "last_modification_time"}, Value: now},
		{Column: clause.Column{Name: "deletion_time"}, Value: now},
	}
	// 定时任务等没有登录用户的场景取不到操作人，只写时间。
	if operatorID, ok := utils.OperatorID(stmt.Context); ok {
		sets = append(sets,
			clause.Assignment{Column: clause.Column{Name: "last_modifier_id"}, Value: operatorID},
			clause.Assignment{Column: clause.Column{Name: "deleter_id"}, Value: operatorID},
		)
	}

	stmt.AddClause(sets)
	stmt.AddClauseIfNotExists(clause.Update{})
	stmt.Clauses[softDeleteTransformKey] = clause.Clause{}
	stmt.Build(stmt.DB.Callback().Update().Clauses...)
}

// softDeleteField 返回模型的 is_deleted 字段，非软删除实体返回 nil。
func softDeleteField(tx *gorm.DB) *schema.Field {
	if tx.Statement.Schema == nil {
		return nil
	}
	field := tx.Statement.Schema.LookUpField("IsDeleted")
	if field == nil {
		return nil
	}
	return field
}

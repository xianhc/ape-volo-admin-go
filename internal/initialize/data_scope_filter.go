package initialize

import (
	"errors"
	"reflect"
	"sync"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/datascope"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// dataScopeClauseKey 标记「本语句已追加过数据权限条件」。
const dataScopeClauseKey = "ape:data_scope"

// dataScopeCreatorColumn 记录创建者的列名。
const dataScopeCreatorColumn = "creator_id"

// dataScopePreloadKey 标记「本语句是某次已过滤查询的关联加载」。
//
// GORM 会把父语句的 Statement.Settings 复制给 Preload 生成的子查询，
// 借此让关联数据跟随主查询的可见性，而不是再按关联表的创建者过滤一遍——
// 否则能看到某个用户，却看不到他的角色和部门。
const dataScopePreloadKey = "ape:data_scope:preloaded"

// dataScopedCache 缓存「模型是否参与数据权限」，避免每次查询都做反射。
var dataScopedCache sync.Map

// RegisterDataScopeFilter 让内嵌 model.DataScoped 的实体按当前用户的可见范围过滤。
//
// 参与过滤的模型内嵌 model.DataScoped，当前用户通过 utils.OperatorID 从上下文读取。
// 解析数据权限自身、登录校验、权限码与菜单树计算需要读全量数据，
// 用 `utils.WithoutDataScope(ctx)` 显式声明。
func RegisterDataScopeFilter(db *gorm.DB) {
	_ = db.Callback().Query().Before("gorm:query").Register("ape:data_scope:query", applyDataScope)
	_ = db.Callback().Update().Before("gorm:update").Register("ape:data_scope:update", applyWriteDataScope)
	_ = db.Callback().Delete().Before("ape:soft_delete:delete").Register("ape:data_scope:delete", applyWriteDataScope)
}

// applyWriteDataScope 拒绝包含不可见记录的整次写入，再为写语句追加权限条件。
// 检查与写入复用当前连接，批量操作不会先修改其中一部分再发现越权。
func applyWriteDataScope(tx *gorm.DB) {
	if tx.Error != nil || tx.Statement == nil || !isDataScopedSchema(tx.Statement.Schema) {
		return
	}
	ctx := tx.Statement.Context
	if utils.DataScopeSkipped(ctx) {
		return
	}
	userId, ok := utils.OperatorID(ctx)
	if !ok {
		return
	}
	scopeCtx := ctx
	if utils.TxFromContext(ctx) == nil || utils.TxSkipped(ctx) {
		scopeCtx = utils.WithTx(ctx, tx)
	}
	scope, err := (&datascope.Service{}).Resolve(scopeCtx, userId)
	if err != nil {
		tx.AddError(err)
		return
	}
	if scope == nil || !scope.Restricted || scope.AllowAll {
		return
	}

	// 主键通常在后续 SQL 构建阶段才补入，此处先补齐以检查准确的目标集合。
	if tx.Statement.Schema != nil {
		_, values := schema.GetIdentityFieldValuesMap(ctx, tx.Statement.ReflectValue, tx.Statement.Schema.PrimaryFields)
		column, keys := schema.ToQueryValues(tx.Statement.Table, tx.Statement.Schema.PrimaryFieldDBNames, values)
		if len(keys) > 0 {
			tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: keys}}})
		}
	}
	condition, hasCondition := tx.Statement.Clauses["WHERE"]
	if !hasCondition {
		tx.AddError(gorm.ErrMissingWhereClause)
		return
	}
	wrapOrConditions(tx)
	condition = tx.Statement.Clauses["WHERE"]
	probe := tx.Session(&gorm.Session{NewDB: true, Context: utils.WithoutDataScope(ctx)}).
		Model(reflect.New(tx.Statement.Schema.ModelType).Interface()).Table(tx.Statement.Table).
		Clauses(condition.Expression)
	column := clause.Column{Table: clause.CurrentTable, Name: dataScopeCreatorColumn}
	if len(scope.UserIds) > 0 {
		values := make([]interface{}, len(scope.UserIds))
		for i, id := range scope.UserIds {
			values[i] = id
		}
		probe = probe.Where(clause.Or(clause.Not(clause.IN{Column: column, Values: values}), clause.Eq{Column: column, Value: nil}))
	}
	var forbidden int64
	if err := probe.Count(&forbidden).Error; err != nil {
		tx.AddError(err)
		return
	}
	if forbidden > 0 {
		tx.AddError(errors.New("数据不存在或您无权查看！"))
		return
	}
	appendCreatorScope(tx, scope)
}

// applyDataScope 给查询追加「可见创建者」条件。
func applyDataScope(tx *gorm.DB) {
	if tx.Error != nil || tx.Statement == nil {
		return
	}
	if !isDataScopedSchema(tx.Statement.Schema) {
		return
	}
	if _, ok := tx.Statement.Clauses[dataScopeClauseKey]; ok {
		return
	}
	// 关联加载不再重复过滤：主查询已经限定了可见记录
	if _, preloaded := tx.Statement.Settings.Load(dataScopePreloadKey); preloaded {
		return
	}

	ctx := tx.Statement.Context
	if utils.DataScopeSkipped(ctx) {
		return
	}

	// 没有登录用户（登录接口、定时任务、启动初始化）时不过滤
	userId, ok := utils.OperatorID(ctx)
	if !ok {
		return
	}

	scope, err := (&datascope.Service{}).Resolve(ctx, userId)
	if err != nil {
		// 取不到范围就报错，不能退化成「不过滤」——那等于把越权变成默认行为
		global.Logger.Error("解析数据权限失败", zap.Error(err))
		tx.AddError(err)
		return
	}
	if scope == nil || !scope.Restricted || scope.AllowAll {
		return
	}
	appendCreatorScope(tx, scope)
	// 让本次查询的 Preload 子查询跟随主记录可见性。
	tx.Statement.Settings.Store(dataScopePreloadKey, true)
}

func appendCreatorScope(tx *gorm.DB, scope *datascope.Snapshot) {
	// 与软删除过滤同理：条件里已有 OR 时必须整体加括号，
	// 否则 `creator_id IN (...) AND a OR b` 会退化成 `(creator_id IN (...) AND a) OR b`。
	wrapOrConditions(tx)

	if len(scope.UserIds) == 0 {
		tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "1 = 0"},
		}})
	} else {
		values := make([]interface{}, 0, len(scope.UserIds))
		for _, id := range scope.UserIds {
			values = append(values, id)
		}
		tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
			clause.IN{
				Column: clause.Column{Table: clause.CurrentTable, Name: dataScopeCreatorColumn},
				Values: values,
			},
		}})
	}
	tx.Statement.Clauses[dataScopeClauseKey] = clause.Clause{}
}

// wrapOrConditions 把已有的 OR 条件整体括起来。
func wrapOrConditions(tx *gorm.DB) {
	c, ok := tx.Statement.Clauses["WHERE"]
	if !ok {
		return
	}
	where, ok := c.Expression.(clause.Where)
	if !ok || len(where.Exprs) == 0 {
		return
	}
	for _, expr := range where.Exprs {
		if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
			where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
			c.Expression = where
			tx.Statement.Clauses["WHERE"] = c
			return
		}
	}
}

// isDataScopedSchema 判断模型是否内嵌了 model.DataScoped。
func isDataScopedSchema(s *schema.Schema) bool {
	if s == nil || s.ModelType == nil {
		return false
	}
	if cached, ok := dataScopedCache.Load(s.ModelType); ok {
		return cached.(bool)
	}

	field, found := s.ModelType.FieldByName("DataScoped")
	result := found && field.Type == reflect.TypeOf(model.DataScoped{})
	dataScopedCache.Store(s.ModelType, result)
	return result
}

package initialize

import (
	"reflect"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// RegisterAuditFiller 注册 ORM 写回调，自动填充审计字段。
//
// 为什么放在 ORM 层而不是每个 handler 里手工设：
//   - 写操作 70+ 处，逐处设置必然遗漏；
//   - 关联表（sys_role_menu / sys_user_role）由 ORM 直接写入，handler 根本感知不到；
//   - 软删除的 DeleterId / DeletionTime 只在「首次置为已删除」时写一次，
//     手工维护很容易在重复更新时被覆盖。
//
// 当前操作人来自 Statement.Context（middleware.JWTAuth 写入请求上下文，
// service 层经 utils.FromContext 把它带到 DB 句柄）。定时任务、启动初始化
// 这类没有登录用户的场景取不到，此时不写用户 ID，只写时间。
func RegisterAuditFiller(db *gorm.DB) {
	_ = db.Callback().Create().Before("gorm:create").Register("ape:audit:fill", fillCreateAudit)
	_ = db.Callback().Update().Before("gorm:update").Register("ape:audit:fill", fillUpdateAudit)
}

func fillCreateAudit(tx *gorm.DB) {
	if tx.Statement == nil {
		return
	}
	operatorID, hasOperator := utils.OperatorID(tx.Statement.Context)
	now := model.Now()

	// Updates(struct) 时 ReflectValue 是 Model（空模板），真正要写入的是 Dest。
	target := tx.Statement.ReflectValue
	if tx.Statement.Dest != nil {
		dv := reflect.ValueOf(tx.Statement.Dest)
		base := dv
		for base.Kind() == reflect.Ptr && !base.IsNil() {
			base = base.Elem()
		}
		// Updates 传的是结构体值时反射不可寻址，字段写不进去。
		// 换成一个可寻址的副本再回写 Dest，GORM 会用副本生成 SET 子句。
		if base.Kind() == reflect.Struct && !base.CanAddr() && modelHasAuditColumns(tx.Statement.Dest) {
			copied := reflect.New(base.Type())
			copied.Elem().Set(base)
			tx.Statement.Dest = copied.Interface()
			dv = copied
		}
		if dv.IsValid() {
			target = dv
		}
	}

	eachModel(target, func(v reflect.Value) {
		base, ok := baseModelOf(v)
		if !ok {
			return
		}
		// 已显式赋值（种子导入等）则不覆盖，只补空值。
		if base.CreationTime.IsZero() {
			base.CreationTime = now
		}
		if hasOperator && base.CreatorId == nil {
			base.CreatorId = &operatorID
		}
	})
}

func fillUpdateAudit(tx *gorm.DB) {
	if tx.Statement == nil {
		return
	}
	operatorID, hasOperator := utils.OperatorID(tx.Statement.Context)
	now := model.Now()

	// Updates(map) 的审计字段只能写进 map，写不进结构体。
	if dest, ok := tx.Statement.Dest.(map[string]interface{}); ok {
		if !modelHasAuditColumns(tx.Statement.Model) {
			return
		}
		dest["last_modification_time"] = now
		if hasOperator {
			dest["last_modifier_id"] = operatorID
		}
		// 置为删除时记录删除人与删除时间；重复删除会刷新为最后一次。
		if deleted, ok := dest["is_deleted"].(bool); ok && deleted {
			dest["deleter_id"] = operatorID
			dest["deletion_time"] = now
		}
		return
	}

	// Updates(struct) 时 ReflectValue 是 Model（空模板），真正要写入的是 Dest。
	target := tx.Statement.ReflectValue
	if tx.Statement.Dest != nil {
		dv := reflect.ValueOf(tx.Statement.Dest)
		base := dv
		for base.Kind() == reflect.Ptr && !base.IsNil() {
			base = base.Elem()
		}
		// Updates 传的是结构体值时反射不可寻址，字段写不进去。
		// 换成一个可寻址的副本再回写 Dest，GORM 会用副本生成 SET 子句。
		if base.Kind() == reflect.Struct && !base.CanAddr() && modelHasAuditColumns(tx.Statement.Dest) {
			copied := reflect.New(base.Type())
			copied.Elem().Set(base)
			tx.Statement.Dest = copied.Interface()
			dv = copied
		}
		if dv.IsValid() {
			target = dv
		}
	}

	eachModel(target, func(v reflect.Value) {
		base, hasBase := baseModelOf(v)
		if hasBase {
			base.LastModificationTime = &now
			if hasOperator {
				id := operatorID
				base.LastModifierId = &id
			}
		}

		// 更新目标里的 DeletionTime 恒为 nil（删除语句只带 IsDeleted 标记），
		// 因此这里每次都会写入，重复删除以最后一次为准，不再回查数据库。
		soft, hasSoft := softDeletedOf(v)
		if !hasSoft || !soft.IsDeleted {
			return
		}
		soft.DeletionTime = &now
		// 删除人取修改人：置为删除本身是一次更新，修改人就是执行删除的人。
		// 定时任务这类没有登录用户的场景取不到，退回上下文里的操作人。
		if hasBase && base.LastModifierId != nil {
			soft.DeleterId = base.LastModifierId
		} else if hasOperator {
			id := operatorID
			soft.DeleterId = &id
		}
	})
}

// eachModel 遍历写入目标，兼容单个结构体、结构体指针与结构体切片。
func eachModel(v reflect.Value, fn func(reflect.Value)) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			eachModel(v.Index(i), fn)
		}
	case reflect.Struct:
		fn(v)
	}
}

func baseModelOf(v reflect.Value) (*model.BaseModel, bool) {
	field := v.FieldByName("BaseModel")
	if !field.IsValid() || !field.CanAddr() {
		return nil, false
	}
	base, ok := field.Addr().Interface().(*model.BaseModel)
	return base, ok
}

func softDeletedOf(v reflect.Value) (*model.SoftDeleted, bool) {
	field := v.FieldByName("SoftDeleted")
	if !field.IsValid() || !field.CanAddr() {
		return nil, false
	}
	soft, ok := field.Addr().Interface().(*model.SoftDeleted)
	return soft, ok
}

// modelHasAuditColumns 判断被更新的模型是否带审计字段。
func modelHasAuditColumns(m any) bool {
	if m == nil {
		return false
	}
	t := reflect.TypeOf(m)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	f, ok := t.FieldByName("BaseModel")
	return ok && f.Type == reflect.TypeOf(model.BaseModel{})
}

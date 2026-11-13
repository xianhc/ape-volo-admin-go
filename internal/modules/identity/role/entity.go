package role

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/menu"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
)

const (
	// DataScopeNone 表示没有可见数据。
	//
	// 数据权限回调会把空范围转换为恒假条件。
	DataScopeNone = 0
	// DataScopeAll 表示可以查看全部数据。
	//
	// 任一角色拥有该范围时不施加创建者过滤。
	DataScopeAll = 1
	// DataScopeMySelf 表示只能查看本人创建的数据。
	//
	// 可见创建者集合只包含当前用户主键。
	DataScopeMySelf = 2
	// DataScopeMyDept 表示可以查看本部门用户创建的数据。
	//
	// 可见创建者集合来自当前部门用户。
	DataScopeMyDept = 3
	// DataScopeMyDeptAndBelow 表示可以查看本部门及下级部门数据。
	//
	// 部门层级查询绕过数据权限以避免树结构被截断。
	DataScopeMyDeptAndBelow = 4
	// DataScopeCustomize 表示按角色关联部门自定义范围。
	//
	// 可见创建者集合来自 sys_role_dept 关联部门用户。
	DataScopeCustomize = 5
)

// Role 表示持久化到数据库的角色。
//
// Menus 与 Departments 为正向关联，Role 不反向依赖 User Entity。
type Role struct {
	model.RootKey
	Name          string                  `gorm:"size:255;uniqueIndex:unique_sys_role_Name;comment:角色名称;not null;"`
	Description   string                  `gorm:"comment:描述"`
	DataScopeType int                     `gorm:"column:data_scope_type;comment:数据权限范围;not null;default:0"`
	AuthCode      string                  `gorm:"column:auth_code;size:255;uniqueIndex:unique_sys_role_AuthCode;comment:角色代码;not null"`
	Menus         []menu.Menu             `gorm:"many2many:sys_role_menu;"`
	Departments   []department.Department `gorm:"many2many:sys_role_dept;"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// DataScopeLabel 返回数据权限范围的中文名称。
//
// 该文本用于角色导出与管理界面展示。
func (role Role) DataScopeLabel() string {
	switch role.DataScopeType {
	case DataScopeAll:
		return "全部"
	case DataScopeMySelf:
		return "本人"
	case DataScopeMyDept:
		return "本部门"
	case DataScopeMyDeptAndBelow:
		return "本部门及以下"
	case DataScopeCustomize:
		return "自定义"
	default:
		return "无"
	}
}

// TableName 返回角色实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Role) TableName() string { return "sys_role" }

// RoleDepartment 表示角色与部门的多对多关联。
//
// 关联实体归属 Role 模块。
type RoleDepartment struct {
	RoleId int64 `gorm:"primaryKey;comment:角色ID"`
	DeptId int64 `gorm:"primaryKey;comment:部门ID"`
}

// TableName 返回角色部门关联表名。
//
// GORM 在数据范围写入时统一使用该名称。
func (RoleDepartment) TableName() string { return "sys_role_dept" }

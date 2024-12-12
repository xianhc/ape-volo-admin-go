package permission

import (
	"go-apevolo/model"
)

type Role struct {
	model.RootKey
	Name        string       `json:"name" gorm:"comment:岗位名称;not null;"`      // 岗位名称
	Level       int          `json:"level" gorm:"comment:角色等级"`               // 角色等级
	Description string       `json:"description" gorm:"comment:描述"`           // 描述
	DataScope   string       `json:"dataScopeType" gorm:"comment:数据权限"`       // 数据权限
	Permission  string       `json:"permission" gorm:"comment:角色代码;not null"` // 角色代码
	Menus       []Menu       `json:"menus"  gorm:"many2many:sys_role_menu;"`  // 菜单列表
	Apis        []Apis       `json:"apis"  gorm:"many2many:sys_role_apis;"`   // 菜单列表
	Departments []Department `json:"depts"  gorm:"many2many:sys_role_dept;"`  // 部门列表
	Users       []User       `json:"users" gorm:"many2many:sys_user_role;"`   //用户列表
	model.BaseModel
	model.SoftDeleted
}

func (Role) TableName() string {
	return "sys_role"
}

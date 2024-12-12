package permission

type RoleMenu struct {
	RoleId int64 `json:"roleId,string" gorm:"comment:角色ID;not null;"` // 角色ID
	MenuId int64 `json:"menuId,string" gorm:"comment:菜单ID;not null;"` // 菜单ID
}

func (RoleMenu) TableName() string {
	return "sys_role_menu"
}

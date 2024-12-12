package permission

type UserRole struct {
	UserId int64 `json:"userId,string" gorm:"comment:用户ID;not null;"` // 用户ID
	RoleId int64 `json:"roleId,string" gorm:"comment:角色ID;not null;"` // 角色ID
}

func (UserRole) TableName() string {
	return "sys_user_role"
}

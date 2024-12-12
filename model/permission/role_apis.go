package permission

type RoleApis struct {
	RoleId int64 `json:"roleId,string" gorm:"comment:角色ID;not null;"`   // 角色ID
	ApisId int64 `json:"apisId,string" gorm:"comment:ApisID;not null;"` // ApisID
}

func (RoleApis) TableName() string {
	return "sys_role_apis"
}

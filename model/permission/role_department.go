package permission

type RoleDepartment struct {
	RoleId int64 `json:"roleId,string" gorm:"comment:角色ID"` // 角色ID
	DeptId int64 `json:"deptId,string" gorm:"comment:部门ID"` // 部门ID
}

func (RoleDepartment) TableName() string {
	return "sys_roles_dept"
}

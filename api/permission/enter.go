package permission

import "go-apevolo/service"

type ApiGroup struct {
	UserApi
	MenuApi
	JobApi
	DeptApi
	RoleApi
	ApisApi
	RolePermissionApi
}

var (
	userService           = service.ServiceGroupApp.PermissionServiceGroup.UserService
	menuService           = service.ServiceGroupApp.PermissionServiceGroup.MenuService
	jobService            = service.ServiceGroupApp.PermissionServiceGroup.JobService
	deptService           = service.ServiceGroupApp.PermissionServiceGroup.DeptService
	roleService           = service.ServiceGroupApp.PermissionServiceGroup.RoleService
	apisService           = service.ServiceGroupApp.PermissionServiceGroup.ApisService
	rolePermissionService = service.ServiceGroupApp.PermissionServiceGroup.RolePermissionService
)

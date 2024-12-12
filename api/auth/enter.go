package auth

import "go-apevolo/service"

type ApiGroup struct {
	AuthorizationApi
}

var (
	userService           = service.ServiceGroupApp.PermissionServiceGroup.UserService
	rolePermissionService = service.ServiceGroupApp.PermissionServiceGroup.RolePermissionService
	queuedService         = service.ServiceGroupApp.QueuedServiceGroup.EmailQueuedService
	tokenBlacklistService = service.ServiceGroupApp.SystemServiceGroup.TokenBlacklistService
)

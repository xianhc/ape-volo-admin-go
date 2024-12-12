package monitor

import "go-apevolo/service"

type ApiGroup struct {
	OnlineUserApi
	AuditLogApi
	ExceptionLogApi
	ServerResourcesApi
}

var (
	onlineUserService   = service.ServiceGroupApp.MonitorServiceGroup.OnlineUserService
	auditLogService     = service.ServiceGroupApp.MonitorServiceGroup.AuditLogService
	exceptionLogService = service.ServiceGroupApp.MonitorServiceGroup.ExceptionLogService
)

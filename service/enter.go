package service

import (
	"go-apevolo/service/message"
	"go-apevolo/service/monitor"
	"go-apevolo/service/permission"
	"go-apevolo/service/queued"
	"go-apevolo/service/system"
)

type ServiceGroup struct {
	SystemServiceGroup     system.ServiceGroup
	PermissionServiceGroup permission.ServiceGroup
	MonitorServiceGroup    monitor.ServiceGroup
	MessageServiceGroup    message.ServiceGroup
	QueuedServiceGroup     queued.ServiceGroup
}

var ServiceGroupApp = new(ServiceGroup)

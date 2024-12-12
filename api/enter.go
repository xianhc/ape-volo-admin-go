package api

import (
	"go-apevolo/api/auth"
	"go-apevolo/api/message/email"
	"go-apevolo/api/monitor"
	"go-apevolo/api/permission"
	"go-apevolo/api/queued"
	"go-apevolo/api/system"
)

type ApiGroup struct {
	SystemApiGroup        system.ApiGroup
	PermissionApiGroup    permission.ApiGroup
	AuthorizationApiGroup auth.ApiGroup
	MonitorApiGroup       monitor.ApiGroup
	MessageApiGroup       email.ApiGroup
	QueuedApiGroup        queued.ApiGroup
}

var ApiGroupApp = new(ApiGroup)

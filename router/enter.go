package router

import (
	"go-apevolo/router/auth"
	"go-apevolo/router/message"
	"go-apevolo/router/monitor"
	"go-apevolo/router/permission"
	"go-apevolo/router/queued"
	"go-apevolo/router/system"
)

type RouterGroup struct {
	System        system.RouterGroup
	Permission    permission.RouterGroup
	Authorization auth.RouterGroup
	Monitor       monitor.RouterGroup
	Message       message.RouterGroup
	Queued        queued.RouterGroup
}

var RouterGroupApp = new(RouterGroup)

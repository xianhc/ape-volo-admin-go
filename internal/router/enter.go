package router

import (
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
	"github.com/xianhc/ape-volo-admin-go/internal/router/auth"
	"github.com/xianhc/ape-volo-admin-go/internal/router/identity"
	"github.com/xianhc/ape-volo-admin-go/internal/router/log"
	"github.com/xianhc/ape-volo-admin-go/internal/router/message"
	"github.com/xianhc/ape-volo-admin-go/internal/router/organization"
	"github.com/xianhc/ape-volo-admin-go/internal/router/queued"
	"github.com/xianhc/ape-volo-admin-go/internal/router/system"
)

type RouterGroup struct {
	System        system.RouterGroup
	Identity      identity.RouterGroup
	Organization  organization.RouterGroup
	Authorization auth.RouterGroup
	Log           log.RouterGroup
	Message       message.RouterGroup
	Queued        queued.RouterGroup
}

var RouterGroupApp = new(RouterGroup)

// InitPrivate 注册全部需要登录的业务路由。
//
// 各模块在注册具体路由时同步声明授权策略；这里仅维护模块挂载顺序，不保存权限映射。
func (group *RouterGroup) InitPrivate(routes *routeauth.Group) {
	group.Authorization.InitAuthorizationRouter(routes)
	group.Identity.InitUserRouter(routes)
	group.Identity.InitMenuRouter(routes)
	group.Identity.InitRoleRouter(routes)
	group.Identity.InitOnlineUserRouter(routes)

	group.Organization.InitDeptRouter(routes)
	group.Organization.InitPositionRouter(routes)

	group.System.InitDictRouter(routes)
	group.System.InitDictDetailRouter(routes)
	group.System.InitSettingRouter(routes)
	group.System.InitAppSecretRouter(routes)
	group.System.InitFileRecordRouter(routes)
	group.System.InitServerResourcesRouter(routes)
	group.System.InitJobRouter(routes)

	group.Log.InitAuditLogRouter(routes)
	group.Log.InitExceptionLogRouter(routes)
	group.Log.InitScheduledJobLogRouter(routes)

	group.Message.InitEmailAccountRouter(routes)
	group.Message.InitMessageTemplateRouter(routes)
	group.Queued.InitEmailQueuedRouter(routes)
}

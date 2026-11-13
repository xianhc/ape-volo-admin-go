package identity

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/onlineuser"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type OnlineUserRouter struct{}

func (s *OnlineUserRouter) InitOnlineUserRouter(Router *routeauth.Group) {
	onlineUserRouter := Router.Group("online", middleware.LogModule("identity", "online"), middleware.OperationRecord())
	onlineUserHandler := onlineuser.Handler{}
	{
		onlineUserRouter.GET("get-list", routeauth.Require(permission.MonitorOnlineList), middleware.LogAction("查询在线用户"), onlineUserHandler.Query)
		onlineUserRouter.DELETE("out", routeauth.Require(permission.MonitorOnlineList), middleware.LogAction("强制用户下线"), onlineUserHandler.DropOut)
		onlineUserRouter.GET("download", routeauth.Require(permission.MonitorOnlineList), middleware.LogAction("导出在线用户"), onlineUserHandler.Download)
	}
}

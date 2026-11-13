package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/serverresources"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type ServerResourcesRouter struct{}

func (s *ServerResourcesRouter) InitServerResourcesRouter(Router *routeauth.Group) {
	serverResourcesRouter := Router.Group("server-resources")
	serverResourcesHandler := serverresources.Handler{}
	{
		serverResourcesRouter.GET("get-list", routeauth.Require(permission.MonitorServerList), serverResourcesHandler.Query)
	}
}

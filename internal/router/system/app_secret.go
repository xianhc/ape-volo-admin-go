package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/appsecret"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type AppSecretRouter struct{}

func (s *AppSecretRouter) InitAppSecretRouter(Router *routeauth.Group) {
	appSecretRouter := Router.Group("app-secret", middleware.LogModule("system", "app-secret"), middleware.OperationRecord())
	appSecretHandler := appsecret.Handler{}
	{
		appSecretRouter.POST("create", routeauth.Require(permission.SysAppSecretAdd), middleware.LogAction("创建应用密钥"), appSecretHandler.Create)
		appSecretRouter.PUT("update", routeauth.Require(permission.SysAppSecretEdit), middleware.LogAction("修改应用密钥"), appSecretHandler.Update)
		appSecretRouter.DELETE("delete", routeauth.Require(permission.SysAppSecretDelete), middleware.LogAction("删除应用密钥"), appSecretHandler.Delete)
		appSecretRouter.GET("get-list", routeauth.Require(permission.SysAppSecretList), middleware.LogAction("查询应用密钥列表"), appSecretHandler.Query)
		appSecretRouter.POST("download", routeauth.Require(permission.SysAppSecretDownload), middleware.LogAction("导出应用密钥"), appSecretHandler.Download)
	}
}

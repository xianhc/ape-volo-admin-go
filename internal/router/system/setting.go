package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/setting"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *routeauth.Group) {
	settingRouter := Router.Group("setting", middleware.LogModule("system", "setting"), middleware.OperationRecord())
	settingHandler := setting.Handler{}
	{
		settingRouter.POST("create", routeauth.Require(permission.SysSettingAdd), middleware.LogAction("创建系统参数"), settingHandler.Create)
		settingRouter.PUT("update", routeauth.Require(permission.SysSettingEdit), middleware.LogAction("修改系统参数"), settingHandler.Update)
		settingRouter.DELETE("delete", routeauth.Require(permission.SysSettingDelete), middleware.LogAction("删除系统参数"), settingHandler.Delete)
		settingRouter.GET("get-list", routeauth.Require(permission.SysSettingList), middleware.LogAction("查询系统参数"), settingHandler.Query)
		settingRouter.POST("download", routeauth.Require(permission.SysSettingDownload), middleware.LogAction("导出系统参数"), settingHandler.Download)
	}
}

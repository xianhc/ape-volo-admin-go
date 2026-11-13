package organization

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/position"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type PositionRouter struct{}

func (s *PositionRouter) InitPositionRouter(Router *routeauth.Group) {
	positionRouter := Router.Group("position", middleware.LogModule("organization", "position"), middleware.OperationRecord())
	positionHandler := position.Handler{}
	{
		positionRouter.POST("create", routeauth.Require(permission.SysPositionAdd), middleware.LogAction("创建岗位"), positionHandler.Create)
		positionRouter.PUT("update", routeauth.Require(permission.SysPositionEdit), middleware.LogAction("修改岗位"), positionHandler.Update)
		positionRouter.DELETE("delete", routeauth.Require(permission.SysPositionDelete), middleware.LogAction("删除岗位"), positionHandler.Delete)
		positionRouter.GET("get-list", routeauth.Require(permission.SysPositionList), middleware.LogAction("查询岗位列表"), positionHandler.Query)
		positionRouter.GET("get-all", routeauth.Require(permission.SysPositionList), middleware.LogAction("查询全部岗位"), positionHandler.All)
		positionRouter.POST("download", routeauth.Require(permission.SysPositionDownload), middleware.LogAction("导出岗位"), positionHandler.Download)
	}
}

package identity

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/menu"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type MenuRouter struct{}

func (s *MenuRouter) InitMenuRouter(Router *routeauth.Group) {
	menuRouter := Router.Group("menu", middleware.LogModule("identity", "menu"), middleware.OperationRecord())
	menuHandler := menu.Handler{}
	{
		menuRouter.POST("create", routeauth.Require(permission.SysMenuAdd), middleware.LogAction("创建菜单"), menuHandler.Create)
		menuRouter.PUT("update", routeauth.Require(permission.SysMenuEdit), middleware.LogAction("修改菜单"), menuHandler.Update)
		menuRouter.DELETE("delete", routeauth.Require(permission.SysMenuDelete), middleware.LogAction("删除菜单"), menuHandler.Delete)
		menuRouter.GET("get-list", routeauth.Require(permission.SysMenuList), middleware.LogAction("查询菜单列表"), menuHandler.Query)
		menuRouter.POST("download", routeauth.Require(permission.SysMenuDownload), middleware.LogAction("导出菜单"), menuHandler.Download)
		menuRouter.POST("build-tree", routeauth.Require(permission.SysMenuList), middleware.LogAction("查询用户菜单树"), menuHandler.Build)
		menuRouter.GET("get-all", routeauth.Require(permission.SysRoleList, rolecode.Admin), middleware.LogAction("查询全部菜单"), menuHandler.All)
	}
}

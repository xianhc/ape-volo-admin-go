package identity

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type RoleRouter struct{}

func (r *RoleRouter) InitRoleRouter(Router *routeauth.Group) {
	roleRouter := Router.Group("role", middleware.LogModule("identity", "role"), middleware.OperationRecord())
	roleHandler := role.Handler{}
	{
		roleRouter.POST("create", routeauth.Require(permission.SysRoleAdd, rolecode.Admin), middleware.LogAction("创建角色"), roleHandler.Create)
		roleRouter.PUT("update", routeauth.Require(permission.SysRoleEdit), middleware.LogAction("修改角色"), roleHandler.Update)
		roleRouter.DELETE("delete", routeauth.Require(permission.SysRoleDelete), middleware.LogAction("删除角色"), roleHandler.Delete)
		roleRouter.GET("get-list", routeauth.Require(permission.SysRoleList), middleware.LogAction("查询角色列表"), roleHandler.Query)
		roleRouter.POST("download", routeauth.Require(permission.SysRoleDownload), middleware.LogAction("导出角色"), roleHandler.Download)
		roleRouter.GET("get-all", routeauth.Require(permission.SysRoleList), middleware.LogAction("查询全部角色"), roleHandler.All)
		roleRouter.GET("get-with-details", routeauth.Require(permission.SysRoleList, rolecode.Admin), middleware.LogAction("查询角色详情"), roleHandler.GetWithDetails)
		roleRouter.PUT("update-role-menu", routeauth.Require(permission.SysRoleEdit, rolecode.Admin), middleware.LogAction("修改角色菜单"), roleHandler.UpdateMenus)
	}
}

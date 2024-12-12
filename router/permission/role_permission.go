package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type RolePermissionRouter struct{}

func (a *RolePermissionRouter) InitRolePermissionRouter(Router *gin.RouterGroup) {
	rolePermissionRouterWithoutRecord := Router.Group("permissions").Use(middleware.OperationRecord())
	rolePermissionApi := api.ApiGroupApp.PermissionApiGroup.RolePermissionApi
	{
		rolePermissionRouterWithoutRecord.GET("menus/query", rolePermissionApi.QueryAllMenus)   // 菜单
		rolePermissionRouterWithoutRecord.GET("apis/query", rolePermissionApi.QueryAllApis)     // 路由
		rolePermissionRouterWithoutRecord.PUT("menus/edit", rolePermissionApi.UpdateRolesMenus) // 更新角色菜单
		rolePermissionRouterWithoutRecord.PUT("apis/edit", rolePermissionApi.UpdateRolesApis)   // 更新角色apis
	}
}

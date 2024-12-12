package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type RoleRouter struct{}

func (r *RoleRouter) InitRoleRouter(Router *gin.RouterGroup) {
	roleRouterWithoutRecord := Router.Group("role").Use(middleware.OperationRecord())
	roleApi := api.ApiGroupApp.PermissionApiGroup.RoleApi
	{
		roleRouterWithoutRecord.POST("create", roleApi.Create)          // 创建
		roleRouterWithoutRecord.PUT("edit", roleApi.Update)             // 编辑
		roleRouterWithoutRecord.DELETE("delete", roleApi.Delete)        // 删除
		roleRouterWithoutRecord.GET("query", roleApi.Query)             // 查询
		roleRouterWithoutRecord.GET("download", roleApi.Download)       // 查询
		roleRouterWithoutRecord.GET("level", roleApi.Level)             // 查询等级
		roleRouterWithoutRecord.GET("all", roleApi.All)                 // 查询全部角色
		roleRouterWithoutRecord.GET("querySingle", roleApi.QuerySingle) // 查看角色
	}
}

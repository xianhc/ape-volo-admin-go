package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type UserRouter struct{}

func (u *UserRouter) InitUserRouter(Router *gin.RouterGroup) {
	userRouterWithoutRecord := Router.Group("user").Use(middleware.OperationRecord())
	userApi := api.ApiGroupApp.PermissionApiGroup.UserApi
	{
		userRouterWithoutRecord.POST("create", userApi.Create)    // 创建
		userRouterWithoutRecord.PUT("edit", userApi.Update)       // 修改
		userRouterWithoutRecord.DELETE("delete", userApi.Delete)  // 删除
		userRouterWithoutRecord.GET("query", userApi.Query)       // 分页获取用户列表
		userRouterWithoutRecord.GET("download", userApi.Download) // 导出
	}
}

package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type ApisRouter struct{}

func (a *ApisRouter) InitApisRouter(Router *gin.RouterGroup) {
	apisRouterWithoutRecord := Router.Group("apis").Use(middleware.OperationRecord())
	apisApi := api.ApiGroupApp.PermissionApiGroup.ApisApi
	{
		apisRouterWithoutRecord.POST("create", apisApi.Create)   // 创建
		apisRouterWithoutRecord.PUT("edit", apisApi.Update)      // 编辑
		apisRouterWithoutRecord.DELETE("delete", apisApi.Delete) // 删除
		apisRouterWithoutRecord.GET("query", apisApi.Query)      // 查询
	}
}

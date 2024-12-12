package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type DictRouter struct{}

func (s *DictRouter) InitDictRouter(Router *gin.RouterGroup) {
	dictRouterWithoutRecord := Router.Group("dict").Use(middleware.OperationRecord())
	dictApi := api.ApiGroupApp.SystemApiGroup.DictApi
	{
		dictRouterWithoutRecord.POST("create", dictApi.Create)    // 创建
		dictRouterWithoutRecord.PUT("edit", dictApi.Update)       // 编辑
		dictRouterWithoutRecord.DELETE("delete", dictApi.Delete)  // 删除
		dictRouterWithoutRecord.GET("query", dictApi.Query)       // 查询
		dictRouterWithoutRecord.GET("download", dictApi.Download) // 导出
	}
}

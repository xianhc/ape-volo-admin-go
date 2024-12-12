package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type DictDetailRouter struct{}

func (s *DictDetailRouter) InitDictDetailRouter(Router *gin.RouterGroup) {
	dictDetailRouterWithoutRecord := Router.Group("dictDetail").Use(middleware.OperationRecord())
	dictDetailApi := api.ApiGroupApp.SystemApiGroup.DictDetailApi
	{
		dictDetailRouterWithoutRecord.POST("create", dictDetailApi.Create)   // 创建
		dictDetailRouterWithoutRecord.PUT("edit", dictDetailApi.Update)      // 编辑
		dictDetailRouterWithoutRecord.DELETE("delete", dictDetailApi.Delete) // 删除
		dictDetailRouterWithoutRecord.GET("query", dictDetailApi.Query)      // 查询
	}
}

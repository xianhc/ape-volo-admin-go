package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	settingRouterWithoutRecord := Router.Group("setting").Use(middleware.OperationRecord())
	settingApi := api.ApiGroupApp.SystemApiGroup.SettingApi
	{
		settingRouterWithoutRecord.POST("create", settingApi.Create)    // 创建
		settingRouterWithoutRecord.PUT("edit", settingApi.Update)       // 编辑
		settingRouterWithoutRecord.DELETE("delete", settingApi.Delete)  // 删除
		settingRouterWithoutRecord.GET("query", settingApi.Query)       // 查询
		settingRouterWithoutRecord.GET("download", settingApi.Download) // 查询
	}
}

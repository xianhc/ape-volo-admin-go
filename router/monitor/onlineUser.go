package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type OnlineUserRouter struct{}

func (s *OnlineUserRouter) InitOnlineUserRouter(Router *gin.RouterGroup) {
	onlineUserRouterWithoutRecord := Router.Group("online").Use(middleware.OperationRecord())
	onlineUserApi := api.ApiGroupApp.MonitorApiGroup.OnlineUserApi
	{
		onlineUserRouterWithoutRecord.GET("query", onlineUserApi.Query)    // 查询
		onlineUserRouterWithoutRecord.DELETE("out", onlineUserApi.DropOut) // 登出
	}
}

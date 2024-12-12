package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
)

type ServerResourcesRouter struct{}

func (s *ServerResourcesRouter) InitServerResourcesRouter(Router *gin.RouterGroup) {
	serverResourcesRouterWithoutRecord := Router.Group("service")
	serverResourcesApi := api.ApiGroupApp.MonitorApiGroup.ServerResourcesApi
	{
		serverResourcesRouterWithoutRecord.GET("resources/info", serverResourcesApi.Query) // 查询
	}
}

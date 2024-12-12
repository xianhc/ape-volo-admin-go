package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
)

type ExceptionLogRouter struct{}

func (s *ExceptionLogRouter) InitExceptionLogRouter(Router *gin.RouterGroup) {
	exceptionLogRouterWithoutRecord := Router.Group("exception")
	exceptionLogApi := api.ApiGroupApp.MonitorApiGroup.ExceptionLogApi
	{
		exceptionLogRouterWithoutRecord.GET("query", exceptionLogApi.Query) // 查询
	}
}

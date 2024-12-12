package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
)

type AuditLogRouter struct{}

func (s *AuditLogRouter) InitAuditLogRouter(Router *gin.RouterGroup) {
	auditLogRouterWithoutRecord := Router.Group("auditing")
	auditLogApi := api.ApiGroupApp.MonitorApiGroup.AuditLogApi
	{
		auditLogRouterWithoutRecord.GET("query", auditLogApi.Query)            // 查询
		auditLogRouterWithoutRecord.GET("current", auditLogApi.QueryByCurrent) // 查询个人
	}
}

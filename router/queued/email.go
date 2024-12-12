package queued

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type EmailQueuedRouter struct{}

func (a *EmailQueuedRouter) InitEmailQueuedRouterWithoutRecord(Router *gin.RouterGroup) {
	emailQueuedRouterWithoutRecord := Router.Group("queued").Use(middleware.OperationRecord())
	emailQueuedApi := api.ApiGroupApp.QueuedApiGroup.EmailQueuedApi
	{
		emailQueuedRouterWithoutRecord.POST("email/create", emailQueuedApi.Create)   // 创建
		emailQueuedRouterWithoutRecord.PUT("email/edit", emailQueuedApi.Update)      // 编辑
		emailQueuedRouterWithoutRecord.DELETE("email/delete", emailQueuedApi.Delete) // 删除
		emailQueuedRouterWithoutRecord.GET("email/query", emailQueuedApi.Query)      // 查询
	}
}

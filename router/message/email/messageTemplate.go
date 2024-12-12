package email

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type MessageTemplateRouter struct{}

func (a *MessageTemplateRouter) InitMessageTemplateRouterWithoutRecord(Router *gin.RouterGroup) {
	messageTemplateRouterWithRecord := Router.Group("email").Use(middleware.OperationRecord())
	messageTemplateApi := api.ApiGroupApp.MessageApiGroup.MessageTemplateApi
	{
		messageTemplateRouterWithRecord.POST("template/create", messageTemplateApi.Create)   // 创建
		messageTemplateRouterWithRecord.PUT("template/edit", messageTemplateApi.Update)      // 编辑
		messageTemplateRouterWithRecord.DELETE("template/delete", messageTemplateApi.Delete) // 删除
		messageTemplateRouterWithRecord.GET("template/query", messageTemplateApi.Query)      // 查询
	}
}

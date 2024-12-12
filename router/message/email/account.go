package email

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type AccountRouter struct{}

func (a *AccountRouter) InitEmailAccountRouterWithoutRecord(Router *gin.RouterGroup) {
	emailAccountRouterWithoutRecord := Router.Group("email").Use(middleware.OperationRecord())
	emailAccountApi := api.ApiGroupApp.MessageApiGroup.AccountApi
	{
		emailAccountRouterWithoutRecord.POST("account/create", emailAccountApi.Create)   // 创建
		emailAccountRouterWithoutRecord.PUT("account/edit", emailAccountApi.Update)      // 编辑
		emailAccountRouterWithoutRecord.DELETE("account/delete", emailAccountApi.Delete) // 删除
		emailAccountRouterWithoutRecord.GET("account/query", emailAccountApi.Query)      // 查询
	}
}

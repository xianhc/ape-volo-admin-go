package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type MenuRouter struct{}

func (s *MenuRouter) InitMenuRouter(Router *gin.RouterGroup) {
	menuRouterWithoutRecord := Router.Group("menu").Use(middleware.OperationRecord())
	menuApi := api.ApiGroupApp.PermissionApiGroup.MenuApi
	{
		menuRouterWithoutRecord.POST("create", menuApi.Create)
		menuRouterWithoutRecord.PUT("edit", menuApi.Update)
		menuRouterWithoutRecord.DELETE("delete", menuApi.Delete)
		menuRouterWithoutRecord.GET("query", menuApi.Query)
		menuRouterWithoutRecord.GET("download", menuApi.Download)
		menuRouterWithoutRecord.GET("superior", menuApi.GetSuperior)
		menuRouterWithoutRecord.GET("child", menuApi.GetChild)
		menuRouterWithoutRecord.GET("build", menuApi.Build)
		menuRouterWithoutRecord.GET("lazy", menuApi.Lazy)
	}
}

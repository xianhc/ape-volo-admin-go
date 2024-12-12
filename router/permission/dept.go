package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type DeptRouter struct{}

func (d *DeptRouter) InitDeptRouter(Router *gin.RouterGroup) {
	deptRouterWithoutRecord := Router.Group("dept").Use(middleware.OperationRecord())
	deptApi := api.ApiGroupApp.PermissionApiGroup.DeptApi
	{
		deptRouterWithoutRecord.POST("create", deptApi.Create)    // 创建
		deptRouterWithoutRecord.PUT("edit", deptApi.Update)       // 编辑
		deptRouterWithoutRecord.DELETE("delete", deptApi.Delete)  // 删除
		deptRouterWithoutRecord.GET("query", deptApi.Query)       // 查询
		deptRouterWithoutRecord.GET("download", deptApi.Download) // 下载
		deptRouterWithoutRecord.GET("superior", deptApi.Superior) // 查询
	}
}

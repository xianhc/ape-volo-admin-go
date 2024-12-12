package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type JobRouter struct{}

func (s *JobRouter) InitJobRouter(Router *gin.RouterGroup) {
	jobRouterWithoutRecord := Router.Group("job").Use(middleware.OperationRecord())
	jobApi := api.ApiGroupApp.PermissionApiGroup.JobApi
	{
		jobRouterWithoutRecord.POST("create", jobApi.Create)    // 创建
		jobRouterWithoutRecord.PUT("edit", jobApi.Update)       // 编辑
		jobRouterWithoutRecord.DELETE("delete", jobApi.Delete)  // 删除
		jobRouterWithoutRecord.GET("query", jobApi.Query)       // 查询
		jobRouterWithoutRecord.GET("queryAll", jobApi.All)      // 查询全部
		jobRouterWithoutRecord.GET("download", jobApi.Download) // 导出
	}
}

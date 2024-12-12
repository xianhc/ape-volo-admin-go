package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type TaskRouter struct{}

func (t *TaskRouter) InitTaskRouter(Router *gin.RouterGroup) {
	taskRouterWithoutRecord := Router.Group("tasks").Use(middleware.OperationRecord())
	taskApi := api.ApiGroupApp.SystemApiGroup.TaskApi
	{
		taskRouterWithoutRecord.POST("create", taskApi.Create)    // 创建
		taskRouterWithoutRecord.PUT("edit", taskApi.Update)       // 编辑
		taskRouterWithoutRecord.DELETE("delete", taskApi.Delete)  // 删除
		taskRouterWithoutRecord.GET("query", taskApi.Query)       // 查询
		taskRouterWithoutRecord.GET("download", taskApi.Download) // 查询
		taskRouterWithoutRecord.PUT("execute", taskApi.Execute)   // 执行
		taskRouterWithoutRecord.PUT("pause", taskApi.Pause)       // 执行
		taskRouterWithoutRecord.PUT("resume", taskApi.Resume)     // 执行
	}
}

package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type FileRecordRouter struct{}

func (s *FileRecordRouter) InitFileRecordRouter(Router *gin.RouterGroup) {
	fileRecordRouterWithoutRecord := Router.Group("storage").Use(middleware.OperationRecord())
	fileRecordApi := api.ApiGroupApp.SystemApiGroup.FileRecordApi
	{
		fileRecordRouterWithoutRecord.POST("upload", fileRecordApi.Upload)    // 创建
		fileRecordRouterWithoutRecord.OPTIONS("upload", fileRecordApi.Upload) // 创建
		fileRecordRouterWithoutRecord.PUT("edit", fileRecordApi.Update)       // 编辑
		fileRecordRouterWithoutRecord.DELETE("delete", fileRecordApi.Delete)  // 删除
		fileRecordRouterWithoutRecord.GET("query", fileRecordApi.Query)       // 查询
		fileRecordRouterWithoutRecord.GET("download", fileRecordApi.Download) // 查询
	}
}

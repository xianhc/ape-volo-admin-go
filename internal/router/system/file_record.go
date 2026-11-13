package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/filerecord"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type FileRecordRouter struct{}

func (s *FileRecordRouter) InitFileRecordRouter(Router *routeauth.Group) {
	fileRecordRouter := Router.Group("file-record", middleware.LogModule("system", "file-record"), middleware.OperationRecord())
	fileRecordHandler := filerecord.Handler{}
	{
		fileRecordRouter.POST("create", routeauth.Require(permission.SysStorageAdd), middleware.LogAction("上传文件"), fileRecordHandler.Upload)
		fileRecordRouter.PUT("update", routeauth.Require(permission.SysStorageEdit), middleware.LogAction("修改文件记录"), fileRecordHandler.Update)
		fileRecordRouter.DELETE("delete", routeauth.Require(permission.SysStorageDelete), middleware.LogAction("删除文件"), fileRecordHandler.Delete)
		fileRecordRouter.GET("get-list", routeauth.Require(permission.SysStorageList), middleware.LogAction("查询文件列表"), fileRecordHandler.Query)
		fileRecordRouter.POST("download", routeauth.Require(permission.SysStorageDownload), middleware.LogAction("导出文件记录"), fileRecordHandler.Download)
	}
}

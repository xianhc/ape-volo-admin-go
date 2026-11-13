package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/dict"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type DictRouter struct{}

func (s *DictRouter) InitDictRouter(Router *routeauth.Group) {
	dictRouter := Router.Group("dict", middleware.LogModule("system", "dict"), middleware.OperationRecord())
	dictHandler := dict.DictHandler{}
	{
		dictRouter.POST("create", routeauth.Require(permission.SysDictAdd), middleware.LogAction("创建字典"), dictHandler.Create)
		dictRouter.PUT("update", routeauth.Require(permission.SysDictEdit), middleware.LogAction("修改字典"), dictHandler.Update)
		dictRouter.DELETE("delete", routeauth.Require(permission.SysDictDelete), middleware.LogAction("删除字典"), dictHandler.Delete)
		dictRouter.GET("get-list", routeauth.Require(permission.SysDictList), middleware.LogAction("查询字典列表"), dictHandler.Query)
		dictRouter.POST("download", routeauth.Require(permission.SysDictDownload), middleware.LogAction("导出字典"), dictHandler.Download)
		dictRouter.GET("get-by-name", routeauth.Require(permission.SysDictList), middleware.LogAction("按名称查询字典"), dictHandler.GetByName)
	}
}

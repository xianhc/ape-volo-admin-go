package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/dict"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type DictDetailRouter struct{}

func (s *DictDetailRouter) InitDictDetailRouter(Router *routeauth.Group) {
	dictDetailRouter := Router.Group("dict-detail", middleware.LogModule("system", "dict-detail"), middleware.OperationRecord())
	detailHandler := dict.DetailHandler{}
	{
		dictDetailRouter.POST("create", routeauth.Require(permission.SysDictAdd), middleware.LogAction("创建字典项"), detailHandler.Create)
		dictDetailRouter.PUT("update", routeauth.Require(permission.SysDictEdit), middleware.LogAction("修改字典项"), detailHandler.Update)
		dictDetailRouter.DELETE("delete", routeauth.Require(permission.SysDictDelete), middleware.LogAction("删除字典项"), detailHandler.Delete)
		dictDetailRouter.GET("get-list", routeauth.Require(permission.SysDictList), middleware.LogAction("查询字典项"), detailHandler.Query)
	}
}

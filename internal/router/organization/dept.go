package organization

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/organization/department"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type DeptRouter struct{}

func (d *DeptRouter) InitDeptRouter(Router *routeauth.Group) {
	deptRouter := Router.Group("department", middleware.LogModule("organization", "department"), middleware.OperationRecord())
	departmentHandler := department.Handler{}
	{
		deptRouter.POST("create", routeauth.Require(permission.SysDeptAdd), middleware.LogAction("创建部门"), departmentHandler.Create)
		deptRouter.PUT("update", routeauth.Require(permission.SysDeptEdit), middleware.LogAction("修改部门"), departmentHandler.Update)
		deptRouter.DELETE("delete", routeauth.Require(permission.SysDeptDelete), middleware.LogAction("删除部门"), departmentHandler.Delete)
		deptRouter.GET("get-list", routeauth.Require(permission.SysDeptList), middleware.LogAction("查询部门列表"), departmentHandler.Query)
		deptRouter.POST("download", routeauth.Require(permission.SysDeptDownload), middleware.LogAction("导出部门"), departmentHandler.Download)
		deptRouter.GET("get-all", routeauth.Require(permission.SysDeptList), middleware.LogAction("查询全部部门"), departmentHandler.All)
	}
}

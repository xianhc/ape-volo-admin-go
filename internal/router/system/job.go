package system

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	jobmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/system/job"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

// JobRouter 注册作业查询、配置和运行时操作的路由。
//
// 私有路由通过 routeauth.Group 同时声明路径和授权策略。
type JobRouter struct{}

// InitJobRouter 注册定时任务查询与编排两组路由。
//
// 查询（get-list / download / get-all）挂在 `job`；
// 配置与运行时的写操作（create / update / delete / execute / pause / resume / stop / restart）
// 挂在 `job-manager`，因为它们在落库之外还要同步内存中的调度器，
// 与只读查询是两类职责。
func (t *JobRouter) InitJobRouter(Router *routeauth.Group) {
	taskHandler := jobmodule.Handler{}

	taskRouter := Router.Group("job", middleware.LogModule("system", "job"), middleware.OperationRecord())
	{
		taskRouter.GET("get-list", routeauth.Require(permission.SysJobList), middleware.LogAction("查询作业列表"), taskHandler.Query)
		taskRouter.POST("download", routeauth.Require(permission.SysJobDownload), middleware.LogAction("导出作业"), taskHandler.Download)
		taskRouter.GET("get-all", routeauth.Require(permission.SysJobList), middleware.LogAction("查询全部作业"), taskHandler.All)
	}

	managerRouter := Router.Group("job-manager", middleware.LogModule("system", "job-manager"), middleware.OperationRecord())
	{
		managerRouter.POST("create", routeauth.Require(permission.SysJobAdd), middleware.LogAction("创建作业"), taskHandler.Create)
		managerRouter.PUT("update", routeauth.Require(permission.SysJobEdit), middleware.LogAction("修改作业"), taskHandler.Update)
		managerRouter.DELETE("delete", routeauth.Require(permission.SysJobDelete), middleware.LogAction("删除作业"), taskHandler.Delete)
		managerRouter.POST("execute", routeauth.Require(permission.SysJobEdit, rolecode.Admin), middleware.LogAction("执行作业"), taskHandler.Execute)
		managerRouter.POST("pause", routeauth.Require(permission.SysJobEdit, rolecode.Admin), middleware.LogAction("暂停作业"), taskHandler.Pause)
		managerRouter.POST("resume", routeauth.Require(permission.SysJobEdit, rolecode.Admin), middleware.LogAction("恢复作业"), taskHandler.Resume)
		managerRouter.POST("stop", routeauth.Require(permission.SysJobEdit), middleware.LogAction("停止作业调度"), taskHandler.Stop)
		managerRouter.POST("restart", routeauth.Require(permission.SysJobEdit), middleware.LogAction("重新调度作业"), taskHandler.Restart)
	}
}

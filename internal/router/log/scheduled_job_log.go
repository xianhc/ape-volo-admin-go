package log

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/scheduledjoblog"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type ScheduledJobLogRouter struct{}

func (t *ScheduledJobLogRouter) InitScheduledJobLogRouter(Router *routeauth.Group) {
	taskLogRouter := Router.Group("job-log", middleware.LogModule("log", "job-log"), middleware.OperationRecord())
	taskLogHandler := scheduledjoblog.Handler{}
	{
		taskLogRouter.GET("get-list", routeauth.Require(permission.SysScheduledJobLogList), middleware.LogAction("查询作业日志"), taskLogHandler.Query)
		taskLogRouter.POST("download", routeauth.Require(permission.SysScheduledJobLogList), middleware.LogAction("导出作业日志"), taskLogHandler.Download)
	}
}

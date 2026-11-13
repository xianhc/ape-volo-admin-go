package log

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/exceptionlog"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type ExceptionLogRouter struct{}

func (s *ExceptionLogRouter) InitExceptionLogRouter(Router *routeauth.Group) {
	exceptionLogRouter := Router.Group("exception-log")
	exceptionLogHandler := exceptionlog.Handler{}
	{
		exceptionLogRouter.GET("get-list", routeauth.Require(permission.LogExceptionList), exceptionLogHandler.Query)
	}
}

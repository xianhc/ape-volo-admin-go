package log

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/operatelog"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type AuditLogRouter struct{}

func (s *AuditLogRouter) InitAuditLogRouter(Router *routeauth.Group) {
	auditLogRouter := Router.Group("operate-log")
	operateLogHandler := operatelog.Handler{}
	{
		auditLogRouter.GET("get-list", routeauth.Require(permission.LogOperateList), operateLogHandler.Query)
		auditLogRouter.GET("get-by-current", routeauth.Require(permission.LogOperateList), operateLogHandler.GetByCurrent)
		auditLogRouter.GET("get-operation-number", routeauth.Require(permission.LogOperateList), operateLogHandler.GetOperationNumber)
	}
}

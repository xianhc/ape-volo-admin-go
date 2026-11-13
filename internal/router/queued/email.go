package queued

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/queuedemail"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type EmailQueuedRouter struct{}

func (a *EmailQueuedRouter) InitEmailQueuedRouter(Router *routeauth.Group) {
	emailQueuedRouter := Router.Group("queued-email", middleware.LogModule("message", "queued-email"), middleware.OperationRecord())
	emailQueuedHandler := queuedemail.Handler{}
	{
		emailQueuedRouter.POST("create", routeauth.Require(permission.MsgQueuedAdd), middleware.LogAction("创建邮件队列项"), emailQueuedHandler.Create)
		emailQueuedRouter.PUT("update", routeauth.Require(permission.MsgQueuedEdit), middleware.LogAction("修改邮件队列项"), emailQueuedHandler.Update)
		emailQueuedRouter.DELETE("delete", routeauth.Require(permission.MsgQueuedDelete), middleware.LogAction("删除邮件队列项"), emailQueuedHandler.Delete)
		emailQueuedRouter.GET("get-list", routeauth.Require(permission.MsgQueuedList), middleware.LogAction("查询邮件队列"), emailQueuedHandler.Query)
	}
}

package email

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailaccount"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type AccountRouter struct{}

func (a *AccountRouter) InitEmailAccountRouter(Router *routeauth.Group) {
	emailAccountRouter := Router.Group("email-account", middleware.LogModule("message", "email-account"), middleware.OperationRecord())
	emailAccountHandler := emailaccount.Handler{}
	{
		emailAccountRouter.POST("create", routeauth.Require(permission.MsgAccountAdd), middleware.LogAction("创建邮件账户"), emailAccountHandler.Create)
		emailAccountRouter.PUT("update", routeauth.Require(permission.MsgAccountEdit), middleware.LogAction("修改邮件账户"), emailAccountHandler.Update)
		emailAccountRouter.DELETE("delete", routeauth.Require(permission.MsgAccountDelete), middleware.LogAction("删除邮件账户"), emailAccountHandler.Delete)
		emailAccountRouter.GET("get-list", routeauth.Require(permission.MsgAccountList), middleware.LogAction("查询邮件账户列表"), emailAccountHandler.Query)
		emailAccountRouter.GET("get-all", routeauth.Require(permission.MsgAccountList), middleware.LogAction("查询全部邮件账户"), emailAccountHandler.All)
	}
}

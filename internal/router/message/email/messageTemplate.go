package email

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailtemplate"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type MessageTemplateRouter struct{}

func (a *MessageTemplateRouter) InitMessageTemplateRouter(Router *routeauth.Group) {
	messageTemplateRouter := Router.Group("email-message-template", middleware.LogModule("message", "email-message-template"), middleware.OperationRecord())
	messageTemplateHandler := emailtemplate.Handler{}
	{
		messageTemplateRouter.POST("create", routeauth.Require(permission.MsgTemplateAdd), middleware.LogAction("创建邮件模板"), messageTemplateHandler.Create)
		messageTemplateRouter.PUT("update", routeauth.Require(permission.MsgTemplateEdit), middleware.LogAction("修改邮件模板"), messageTemplateHandler.Update)
		messageTemplateRouter.DELETE("delete", routeauth.Require(permission.MsgTemplateDelete), middleware.LogAction("删除邮件模板"), messageTemplateHandler.Delete)
		messageTemplateRouter.GET("get-list", routeauth.Require(permission.MsgTemplateList), middleware.LogAction("查询邮件模板列表"), messageTemplateHandler.Query)
	}
}

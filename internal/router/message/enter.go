package message

import "github.com/xianhc/ape-volo-admin-go/internal/router/message/email"

type RouterGroup struct {
	email.AccountRouter
	email.MessageTemplateRouter
}

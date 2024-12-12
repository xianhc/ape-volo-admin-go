package message

import "go-apevolo/router/message/email"

type RouterGroup struct {
	email.AccountRouter
	email.MessageTemplateRouter
}

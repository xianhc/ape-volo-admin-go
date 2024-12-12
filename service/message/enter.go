package message

import "go-apevolo/service/message/email"

type ServiceGroup struct {
	email.AccountService
	email.MessageTemplateService
}

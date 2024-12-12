package email

import (
	"go-apevolo/service"
)

type ApiGroup struct {
	AccountApi
	MessageTemplateApi
}

var (
	emailAccountService         = service.ServiceGroupApp.MessageServiceGroup.AccountService
	emailMessageTemplateService = service.ServiceGroupApp.MessageServiceGroup.MessageTemplateService
)

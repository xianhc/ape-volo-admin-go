package queued

import (
	"go-apevolo/service"
)

type ApiGroup struct {
	EmailQueuedApi
}

var (
	emailQueuedService = service.ServiceGroupApp.QueuedServiceGroup.EmailQueuedService
)

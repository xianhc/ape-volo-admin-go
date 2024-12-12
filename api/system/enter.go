package system

import "go-apevolo/service"

type ApiGroup struct {
	DictApi
	DictDetailApi
	SettingApi
	AppSecretApi
	FileRecordApi
	TaskApi
}

var (
	dictService       = service.ServiceGroupApp.SystemServiceGroup.DictService
	dictDetailService = service.ServiceGroupApp.SystemServiceGroup.DictDetailService
	settingService    = service.ServiceGroupApp.SystemServiceGroup.SettingService
	appSecretService  = service.ServiceGroupApp.SystemServiceGroup.AppSecretService
	fileRecordService = service.ServiceGroupApp.SystemServiceGroup.FileRecordService
	taskService       = service.ServiceGroupApp.SystemServiceGroup.TaskService
)

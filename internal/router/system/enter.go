package system

type RouterGroup struct {
	DictRouter
	DictDetailRouter
	SettingRouter
	AppSecretRouter
	FileRecordRouter
	ServerResourcesRouter
	JobRouter
}

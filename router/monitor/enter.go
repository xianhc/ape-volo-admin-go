package monitor

type RouterGroup struct {
	OnlineUserRouter
	AuditLogRouter
	ExceptionLogRouter
	ServerResourcesRouter
}

package log

type RouterGroup struct {
	AuditLogRouter
	ExceptionLogRouter
	ScheduledJobLogRouter
}

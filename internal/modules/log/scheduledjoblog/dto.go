package scheduledjoblog

// Execution 表示记录执行日志所需的任务快照。
//
// 只传递日志需要的字段，日志模块不依赖任务配置或执行入口。
type Execution struct {
	JobId        int64
	JobName      string
	JobGroup     string
	AssemblyName string
	ClassName    string
	Cron         string
	RunParams    string
}

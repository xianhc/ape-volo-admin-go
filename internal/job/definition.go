package job

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Definition 表示任务执行入口使用的调度配置快照。
//
// 该类型不进入 HTTP 响应，执行入口不依赖任务配置模块的数据库实体。
type Definition struct {
	Id                int64
	JobName           string
	JobGroup          string
	Cron              string
	ClassName         string
	AssemblyName      string
	Description       string
	Principal         string
	AlertEmail        string
	PauseAfterFailure bool
	StartTime         *model.Time
	EndTime           *model.Time
	TriggerType       int32
	IntervalSecond    *int32
	CycleRunTimes     *int32
	Enabled           bool
	RunParams         string
}

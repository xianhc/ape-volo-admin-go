package scheduledjoblog

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// ScheduledJobLog 表示持久化到数据库的调度任务执行日志。
//
// 实体保存任务快照、执行耗时、结果和异常详情，不直接作为 HTTP 响应返回。
type ScheduledJobLog struct {
	model.RootKey
	JobId             int64                 `gorm:"column:job_id;comment:任务ID"`
	JobName           string                `gorm:"comment:任务名称"`
	JobGroup          string                `gorm:"size:255;comment:任务分组"`
	AssemblyName      string                `gorm:"comment:程序集名称"`
	ClassName         string                `gorm:"comment:任务所在类"`
	Cron              string                `gorm:"comment:cron 表达式"`
	RunParams         model.CustomFieldText `gorm:"comment:执行传参"`
	ExceptionDetail   model.CustomFieldText `gorm:"comment:异常详情"`
	ExecutionDuration int64                 `gorm:"default:0;comment:执行耗时(毫秒)"`
	IsSuccess         bool                  `gorm:"comment:是否成功"`
	CreatorName       string                `gorm:"comment:创建者账号"`
	model.BaseModel
	model.SoftDeleted
}

// TableName 返回调度任务日志实体的月表基名。
//
// 实际读写由仓储显式选择月份。
func (ScheduledJobLog) TableName() string { return "log_job" }

// LogID 返回生产端已经确定的记录 ID。
//
// 用于写入错误摘要，不输出作业参数。
func (l ScheduledJobLog) LogID() int64 { return l.Id }

// LogTime 返回作业执行结束时确定的创建时刻。
//
// 消费者据此选月，不按消费时间改写。
func (l ScheduledJobLog) LogTime() time.Time { return l.CreationTime.Time }

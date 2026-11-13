package job

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Job 表示持久化到数据库的调度任务配置。
//
// TriggerState 是运行时派生字段，不写入数据库；Entity 不直接作为 HTTP 响应返回。
type Job struct {
	model.RootKey
	JobName           string      `gorm:"comment:任务名称"`
	JobGroup          string      `gorm:"comment:任务分组"`
	Cron              string      `gorm:"comment:cron 表达式"`
	ClassName         string      `gorm:"comment:任务所在类"`
	AssemblyName      string      `gorm:"comment:程序集名称"`
	Description       string      `gorm:"comment:任务描述"`
	Principal         string      `gorm:"comment:任务负责人"`
	AlertEmail        string      `gorm:"comment:告警邮箱"`
	PauseAfterFailure bool        `gorm:"comment:任务失败后是否暂停"`
	StartTime         *model.Time `gorm:"comment:开始时间"`
	EndTime           *model.Time `gorm:"comment:结束时间"`
	TriggerType       int32       `gorm:"comment:触发器类型（1、cron 2、simple）"`
	IntervalSecond    *int32      `gorm:"comment:执行间隔时间, 秒为单位"`
	CycleRunTimes     *int32      `gorm:"comment:循环执行次数"`
	Enabled           bool        `gorm:"comment:是否启动"`
	RunParams         string      `gorm:"comment:执行传参"`
	TriggerState      int32       `gorm:"-"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回调度任务实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Job) TableName() string { return "sys_job" }

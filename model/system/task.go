package system

import (
	"go-apevolo/model"
	"time"
)

type Task struct {
	model.RootKey
	TaskName          string     `json:"taskName" gorm:"comment:任务名称"`                      // 任务名称
	TaskGroup         string     `json:"taskGroup" gorm:"comment:任务分组"`                     // 任务分组
	Cron              string     `json:"cron"  gorm:"comment:cron 表达式"`                     // 表达式
	ClassName         string     `json:"className" gorm:"comment:任务所在类"`                    // 任务所在类
	AssemblyName      string     `json:"assemblyName" gorm:"comment:程序集名称"`                 // 程序集名称
	Description       string     `json:"description" gorm:"comment:任务描述"`                   // 任务描述
	Principal         string     `json:"principal" gorm:"comment:任务负责人"`                    // 任务负责人
	AlertEmail        string     `json:"alertEmail" gorm:"comment:告警邮箱"`                    // 告警邮箱
	PauseAfterFailure bool       `json:"pauseAfterFailure" gorm:"comment:任务失败后是否暂停"`        // 任务失败后是否暂停
	RunTimes          int32      `json:"runTimes" gorm:"comment:执行次数"`                      // 执行次数
	StartTime         *time.Time `json:"startTime" gorm:"comment:开始时间"`                     // 开始时间
	EndTime           *time.Time `json:"endTime" gorm:"comment:结束时间"`                       // 结束时间
	TriggerType       int32      `json:"triggerType" gorm:"comment:触发器类型（0、simple 1、cron）"` // 触发器类型（0、simple 1、cron）
	IntervalSecond    *int32     `json:"intervalSecond" gorm:"comment:执行间隔时间, 秒为单位"`        // 执行间隔时间, 秒为单位
	CycleRunTimes     *int32     `json:"cycleRunTimes" gorm:"comment:循环执行次数"`               // 循环执行次数
	IsEnable          bool       `json:"isEnable" gorm:"comment:是否启动"`                      // 是否启动
	RunParams         string     `json:"runParams" gorm:"comment:执行传参"`                     // 执行传参
	TriggerStatus     string     `json:"triggerStatus" gorm:"-"`                            // 触发器状态
	model.BaseModel
	model.SoftDeleted
}

func (Task) TableName() string {
	return "sys_quartz_job"
}

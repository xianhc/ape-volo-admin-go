package system

import (
	"go-apevolo/model"
)

type TaskLog struct {
	model.RootKey
	TaskId            int64  `json:"TaskId" gorm:"comment:任务ID"`            // 任务ID
	TaskName          string `json:"taskName" gorm:"comment:任务名称"`          // 任务名称
	TaskGroup         string `json:"taskGroup" gorm:"comment:任务分组"`         // 任务组
	AssemblyName      string `json:"assemblyName" gorm:"comment:任务描述"`      // 命名空间
	ClassName         string `json:"className" gorm:"comment:任务所在类"`        // 类名称
	Cron              string `json:"cron"  gorm:"comment:cron 表达式"`         // Cron表达式
	RunParams         string `json:"runParams" gorm:"comment:执行传参"`         // 执行传参
	ExceptionDetail   string `json:"exceptionDetail" gorm:"comment:执行传参"`   // 异常详情
	ExecutionDuration int64  `json:"ExecutionDuration" gorm:"comment:执行传参"` // 执行耗时
	IsSuccess         bool   `json:"isSuccess" gorm:"comment:执行传参"`         // 是否成功
	model.BaseModel
	model.SoftDeleted
}

func (TaskLog) TableName() string {
	return "sys_quartz_job_log"
}

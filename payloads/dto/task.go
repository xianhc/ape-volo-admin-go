package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateTaskDto struct {
	model.RootKey
	TaskName          string     `json:"taskName" validate:"required"`     // 任务名称
	TaskGroup         string     `json:"taskGroup" validate:"required"`    // 任务分组
	Cron              string     `json:"cron" `                            // 表达式
	ClassName         string     `json:"className" validate:"required"`    // 任务所在类
	AssemblyName      string     `json:"assemblyName" validate:"required"` // 程序集名称
	Description       string     `json:"description" `                     // 任务描述
	Principal         string     `json:"principal"`                        // 任务负责人
	AlertEmail        string     `json:"alertEmail" `                      // 告警邮箱
	PauseAfterFailure bool       `json:"pauseAfterFailure" `               // 任务失败后是否暂停
	RunTimes          int32      `json:"runTimes" `                        // 执行次数
	StartTime         *time.Time `json:"startTime" `                       // 开始时间
	EndTime           *time.Time `json:"endTime" `                         // 结束时间
	TriggerType       int32      `json:"triggerType" `                     // 触发器类型（0、simple 1、cron）
	IntervalSecond    *int32     `json:"intervalSecond" `                  // 执行间隔时间, 秒为单位
	CycleRunTimes     *int32     `json:"cycleRunTimes" `                   // 循环执行次数
	IsEnable          bool       `json:"isEnable" `                        // 是否启动
	RunParams         string     `json:"runParams"`                        // 执行传参
	model.BaseModel
}

func (req *CreateUpdateTaskDto) Generate(model *system.Task) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.TaskName = req.TaskName
	model.TaskGroup = req.TaskGroup
	model.Cron = req.Cron
	model.ClassName = req.ClassName
	model.AssemblyName = req.AssemblyName
	model.Description = req.Description
	model.Principal = req.Principal
	model.AlertEmail = req.AlertEmail
	model.PauseAfterFailure = req.PauseAfterFailure
	model.RunTimes = req.RunTimes
	model.StartTime = req.StartTime
	model.EndTime = req.EndTime
	model.TriggerType = req.TriggerType
	model.IntervalSecond = req.IntervalSecond
	model.CycleRunTimes = req.CycleRunTimes
	model.IsEnable = req.IsEnable
	model.RunParams = req.RunParams
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type TaskQueryCriteria struct {
	TaskName   string   `json:"taskName" form:"taskName"`
	CreateTime []string `json:"createTime" form:"createTime"`
	request.Pagination
}

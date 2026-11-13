package job

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// JobResponse 表示调度任务接口的返回模型。
//
// TriggerState 由运行时调度器计算，数据库 Entity 不进入 HTTP 输出。
type JobResponse struct {
	Id                   int64       `json:"id,string"`
	JobName              string      `json:"jobName"`
	JobGroup             string      `json:"jobGroup"`
	Cron                 string      `json:"cron"`
	ClassName            string      `json:"className"`
	AssemblyName         string      `json:"assemblyName"`
	Description          string      `json:"description"`
	Principal            string      `json:"principal"`
	AlertEmail           string      `json:"alertEmail"`
	PauseAfterFailure    bool        `json:"pauseAfterFailure"`
	StartTime            *model.Time `json:"startTime"`
	EndTime              *model.Time `json:"endTime"`
	TriggerType          int32       `json:"triggerType"`
	IntervalSecond       *int32      `json:"intervalSecond"`
	CycleRunTimes        *int32      `json:"cycleRunTimes"`
	Enabled              bool        `json:"enabled"`
	RunParams            string      `json:"runParams"`
	TriggerState         int32       `json:"triggerState"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将调度任务实体转换为接口返回模型。
//
// triggerState 参数允许 Service 注入当前运行时状态。
func fromEntity(task Job, triggerState int32) JobResponse {
	return JobResponse{Id: task.Id, JobName: task.JobName, JobGroup: task.JobGroup, Cron: task.Cron, ClassName: task.ClassName, AssemblyName: task.AssemblyName, Description: task.Description, Principal: task.Principal, AlertEmail: task.AlertEmail, PauseAfterFailure: task.PauseAfterFailure, StartTime: task.StartTime, EndTime: task.EndTime, TriggerType: task.TriggerType, IntervalSecond: task.IntervalSecond, CycleRunTimes: task.CycleRunTimes, Enabled: task.Enabled, RunParams: task.RunParams, TriggerState: triggerState, CreatorId: task.CreatorId, CreationTime: task.CreationTime, LastModifierId: task.LastModifierId, LastModificationTime: task.LastModificationTime, DeleterId: task.DeleterId, DeletionTime: task.DeletionTime}
}

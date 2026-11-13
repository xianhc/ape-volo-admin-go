package scheduledjoblog

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// ScheduledJobLogResponse 表示调度任务日志接口的返回模型。
//
// ExecutionDuration 使用 camelCase JSON 字段名，与统一响应契约一致。
type ScheduledJobLogResponse struct {
	Id                   int64       `json:"id,string"`
	JobId                int64       `json:"jobId,string"`
	JobName              string      `json:"jobName"`
	JobGroup             string      `json:"jobGroup"`
	AssemblyName         string      `json:"assemblyName"`
	ClassName            string      `json:"className"`
	Cron                 string      `json:"cron"`
	RunParams            string      `json:"runParams"`
	ExceptionDetail      string      `json:"exceptionDetail"`
	ExecutionDuration    int64       `json:"executionDuration"`
	IsSuccess            bool        `json:"isSuccess"`
	CreatorName          string      `json:"creatorName"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将调度任务日志实体转换为接口返回模型。
//
// 转换使用显式字段映射并隐藏软删除标记。
func fromEntity(log ScheduledJobLog) ScheduledJobLogResponse {
	return ScheduledJobLogResponse{Id: log.Id, JobId: log.JobId, JobName: log.JobName, JobGroup: log.JobGroup, AssemblyName: log.AssemblyName, ClassName: log.ClassName, Cron: log.Cron, RunParams: string(log.RunParams), ExceptionDetail: string(log.ExceptionDetail), ExecutionDuration: log.ExecutionDuration, IsSuccess: log.IsSuccess, CreatorName: log.CreatorName, CreatorId: log.CreatorId, CreationTime: log.CreationTime, LastModifierId: log.LastModifierId, LastModificationTime: log.LastModificationTime, DeleterId: log.DeleterId, DeletionTime: log.DeletionTime}
}

// fromEntities 将调度任务日志实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(logs []ScheduledJobLog) []ScheduledJobLogResponse {
	result := make([]ScheduledJobLogResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, fromEntity(log))
	}
	return result
}

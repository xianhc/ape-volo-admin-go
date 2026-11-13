package job

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateJobRequest 表示创建或修改调度任务的请求参数。
//
// 触发器相关组合约束由 Service 根据 TriggerType 校验。
type CreateAndUpdateJobRequest struct {
	Id                int64       `json:"id,string"`
	JobName           string      `json:"jobName" validate:"required"`
	JobGroup          string      `json:"jobGroup" validate:"required"`
	Cron              string      `json:"cron"`
	ClassName         string      `json:"className" validate:"required"`
	AssemblyName      string      `json:"assemblyName" validate:"required"`
	Description       string      `json:"description"`
	Principal         string      `json:"principal"`
	AlertEmail        string      `json:"alertEmail"`
	PauseAfterFailure bool        `json:"pauseAfterFailure"`
	StartTime         *model.Time `json:"startTime"`
	EndTime           *model.Time `json:"endTime"`
	TriggerType       int32       `json:"triggerType"`
	IntervalSecond    *int32      `json:"intervalSecond"`
	CycleRunTimes     *int32      `json:"cycleRunTimes"`
	Enabled           bool        `json:"enabled"`
	RunParams         string      `json:"runParams"`
}

// QueryJobRequest 表示调度任务列表与导出的筛选参数。
//
// 支持任务名称和创建时间范围筛选。
type QueryJobRequest struct {
	request.DateRangePagination[string]
	JobName string `json:"jobName" form:"jobName"`
}

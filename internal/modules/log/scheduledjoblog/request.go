package scheduledjoblog

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// QueryScheduledJobLogRequest 表示调度任务日志的筛选参数。
//
// 支持任务名称模糊匹配、任务组和执行结果筛选。
type QueryScheduledJobLogRequest struct {
	request.Pagination
	JobName   string `json:"jobName" form:"jobName"`
	JobGroup  string `json:"jobGroup" form:"jobGroup"`
	IsSuccess *bool  `json:"isSuccess" form:"isSuccess"`
}

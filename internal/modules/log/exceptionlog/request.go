package exceptionlog

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// QueryExceptionLogRequest 表示异常日志列表的筛选参数。
//
// CreateTime 接受起止时间，分页与排序字段保持既有接口契约。
type QueryExceptionLogRequest struct {
	request.DateRangePagination[time.Time]
	KeyWords string `json:"keyWords" form:"keyWords"`
}

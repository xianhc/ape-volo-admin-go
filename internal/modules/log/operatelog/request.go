package operatelog

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// QueryOperateLogRequest 表示操作日志列表的筛选参数。
//
// CreateTime 接受起止时间，分页与排序字段保持既有接口契约。
type QueryOperateLogRequest struct {
	request.DateRangePagination[time.Time]
	KeyWords string `json:"keyWords" form:"keyWords"`
}

// QueryCurrentRequest 表示当前用户操作日志的分页参数。
//
// 当前用户主键从认证上下文读取，不接受客户端传入。
type QueryCurrentRequest struct {
	request.Pagination
}

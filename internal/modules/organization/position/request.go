package position

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdatePositionRequest 表示创建或修改岗位的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdatePositionRequest struct {
	Id      int64  `json:"id,string"`
	Name    string `json:"name" validate:"required"`
	Sort    int    `json:"sort" validate:"required,min=1,max=999"`
	Enabled bool   `json:"enabled"`
}

// QueryPositionRequest 表示岗位列表与导出的筛选参数。
//
// JobName 字段名为既有接口契约，实际筛选岗位名称。
type QueryPositionRequest struct {
	request.DateRangePagination[time.Time]
	JobName string `json:"jobName" form:"jobName"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

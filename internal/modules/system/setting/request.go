package setting

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateSettingRequest 表示创建或修改系统设置的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdateSettingRequest struct {
	Id          int64  `json:"id,string" form:"id"`
	Name        string `json:"name" form:"name" validate:"required"`
	Value       string `json:"value" form:"value" validate:"required"`
	Enabled     bool   `json:"enabled" form:"enabled"`
	Description string `json:"description" form:"description"`
}

// QuerySettingRequest 表示系统设置列表与导出的筛选参数。
//
// CreateTime 为既有请求契约字段；当前查询行为不按该字段过滤。
type QuerySettingRequest struct {
	request.DateRangePagination[time.Time]
	KeyWords string `json:"keyWords" form:"keyWords"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
}

package emailtemplate

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateTemplateRequest 表示创建或修改邮件模板的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdateTemplateRequest struct {
	Id                int64  `json:"id,string"`
	Name              string `json:"name" validate:"required"`
	BccEmailAddresses string `json:"bccEmailAddresses"`
	Subject           string `json:"subject" validate:"required"`
	Body              string `json:"body" validate:"required"`
	Enabled           bool   `json:"enabled"`
	EmailAccountId    int64  `json:"emailAccountId"`
}

// QueryTemplateRequest 表示邮件模板列表的筛选参数。
//
// 支持模板名称、激活状态和创建时间范围筛选。
type QueryTemplateRequest struct {
	request.DateRangePagination[time.Time]
	Name    string `json:"name" form:"name"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

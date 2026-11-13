package emailaccount

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateAccountRequest 表示创建或修改邮件账户的请求参数。
//
// 创建时 Password 使用传入值；修改时为空则保留原密码。
type CreateAndUpdateAccountRequest struct {
	Id                    int64  `json:"id,string"`
	Email                 string `json:"email" validate:"required,email"`
	DisplayName           string `json:"displayName" validate:"required"`
	Host                  string `json:"host" validate:"required"`
	Port                  int32  `json:"port"`
	UserName              string `json:"userName" validate:"required"`
	Password              string `json:"password"`
	EnableSsl             bool   `json:"enableSsl"`
	UseDefaultCredentials bool   `json:"useDefaultCredentials"`
}

// QueryAccountRequest 表示邮件账户列表与导出的筛选参数。
//
// 支持邮箱、显示名称、用户名和创建时间范围筛选。
type QueryAccountRequest struct {
	request.DateRangePagination[time.Time]
	Email       string `json:"email" form:"email"`
	DisplayName string `json:"displayName" form:"displayName"`
	UserName    string `json:"userName" form:"userName"`
}

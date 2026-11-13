package appsecret

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateAppSecretRequest 表示创建或修改应用密钥的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由服务生成主键、应用 ID 和密钥。
type CreateAndUpdateAppSecretRequest struct {
	Id      int64  `json:"id,string" form:"id"`
	AppName string `json:"appName" form:"appName" validate:"required"`
	Remark  string `json:"remark" form:"remark"`
}

// QueryAppSecretRequest 表示应用密钥列表与导出的筛选参数。
//
// 分页查询使用 PageIndex 和 PageSize，导出只使用筛选与排序字段。
type QueryAppSecretRequest struct {
	request.DateRangePagination[string]
	KeyWords string `json:"keyWords" form:"keyWords"`
}

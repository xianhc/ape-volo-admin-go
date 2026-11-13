package dict

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// DictReference 表示字典明细请求中的所属字典引用。
//
// 当前接口只需要字典主键，不接收字典的其他字段。
type DictReference struct {
	Id int64 `json:"id,string"`
}

// CreateAndUpdateDictDetailRequest 表示创建或修改字典明细的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdateDictDetailRequest struct {
	Id       int64         `json:"id,string"`
	Label    string        `json:"label" validate:"required"`
	Value    string        `json:"value" validate:"required"`
	DictSort int           `json:"dictSort" validate:"required,min=1,max=999"`
	Dict     DictReference `json:"dict"`
}

// CreateAndUpdateDictRequest 表示创建或修改字典的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdateDictRequest struct {
	Id          int64  `json:"id,string"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

// QueryDictRequest 表示字典列表与导出的筛选参数。
//
// 分页查询使用 PageIndex 和 PageSize，导出忽略分页字段。
type QueryDictRequest struct {
	request.Pagination
	KeyWords string `json:"keyWords" form:"keyWords"`
}

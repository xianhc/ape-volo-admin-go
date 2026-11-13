package filerecord

import (
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateFileRecordRequest 表示上传文件时创建记录的请求参数。
//
// 文件自身通过 multipart 表单传入，该结构只承载文件描述和兼容主键。
type CreateAndUpdateFileRecordRequest struct {
	Id          int64  `json:"id,string" form:"id"`
	Description string `json:"description" form:"description" validate:"required"`
}

// QueryFileRecordRequest 表示文件记录列表与导出的筛选参数。
//
// 分页查询使用 PageIndex 和 PageSize，导出忽略分页字段。
type QueryFileRecordRequest struct {
	request.DateRangePagination[string]
	KeyWords string `json:"keyWords" form:"keyWords"`
}

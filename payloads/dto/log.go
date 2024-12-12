package dto

import (
	"go-apevolo/payloads/request"
	"time"
)

type LogQueryCriteria struct {
	KeyWords   string      `json:"keyWords" form:"keyWords"`
	CreateTime []time.Time `json:"createTime" form:"createTime"`
	request.Pagination
}

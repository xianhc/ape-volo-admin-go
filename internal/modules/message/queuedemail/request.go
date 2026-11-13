package queuedemail

import (
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
)

// CreateAndUpdateQueuedEmailRequest 表示创建或修改队列邮件的请求参数。
//
// 创建时 Id 为兼容已有调用保留；未提供时由 Service 生成主键。
type CreateAndUpdateQueuedEmailRequest struct {
	Id             int64       `json:"id,string"`
	From           string      `json:"from"`
	FromName       string      `json:"fromName"`
	To             string      `json:"to" validate:"required"`
	ToName         string      `json:"toName"`
	ReplyTo        string      `json:"replyTo"`
	ReplyToName    string      `json:"replyToName"`
	Priority       int32       `json:"priority"`
	Cc             string      `json:"cc"`
	Bcc            string      `json:"bcc"`
	Subject        string      `json:"subject"`
	Body           string      `json:"body" validate:"required"`
	SentTries      int32       `json:"sentTries"`
	SendTime       *model.Time `json:"sendTime"`
	EmailAccountId int64       `json:"emailAccountId"`
}

// QueryQueuedEmailRequest 表示队列邮件列表的筛选参数。
//
// 支持主键、重试次数、发送账户、收件人、发送状态和创建时间筛选。
type QueryQueuedEmailRequest struct {
	request.DateRangePagination[time.Time]
	Id             *int64 `json:"id" form:"id"`
	MaxTries       int32  `json:"maxTries" form:"maxTries"`
	EmailAccountId *int64 `json:"emailAccountId" form:"emailAccountId"`
	To             string `json:"to" form:"to"`
	IsSend         *bool  `json:"isSend" form:"isSend"`
}

package queuedemail

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Delivery 表示邮件发送任务需要的队列数据。
//
// 该 DTO 不进入 HTTP 响应，仅在队列 Service 与定时任务之间传递。
type Delivery struct {
	Id             int64
	From           string
	FromName       string
	To             string
	Subject        string
	Body           string
	Priority       int32
	SentTries      int32
	IsSend         bool
	SendTime       *model.Time
	EmailAccountId int64
}

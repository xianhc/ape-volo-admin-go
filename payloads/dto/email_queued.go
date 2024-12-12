package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/queued"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateEmailQueuedDto struct {
	model.RootKey
	To             string     `json:"to" validate:"required"`   // 收件邮箱
	ToName         string     `json:"toName"`                   // 收件人名称
	ReplyTo        string     `json:"replyTo"`                  // 回复邮箱
	ReplyToName    string     `json:"replyToName" `             // 回复人名称
	Priority       int32      `json:"priority"`                 // 优先级
	Cc             string     `json:"cc"`                       // 抄送
	Bcc            string     `json:"bcc"`                      // 密件抄送
	Subject        string     `json:"subject"`                  // 标题
	Body           string     `json:"body" validate:"required"` // 内容
	SentTries      int32      `json:"sentTries"`                // 发送上限次数
	SendTime       *time.Time `json:"sendTime"`                 // 发送时间
	EmailAccountId int64      `json:"emailAccountId"`           // 发件邮箱ID
	model.BaseModel
}

func (req *CreateUpdateEmailQueuedDto) Generate(model *queued.Email) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.To = req.To
	model.ToName = req.ToName
	model.ReplyTo = req.ReplyTo
	model.ReplyToName = req.ReplyToName
	model.Priority = req.Priority
	model.Cc = req.Cc
	model.Bcc = req.Bcc
	model.Subject = req.Subject
	model.Body = req.Body
	model.SentTries = req.SentTries
	model.SendTime = req.SendTime
	model.EmailAccountId = req.EmailAccountId
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type EmailQueuedQueryCriteria struct {
	Id             *int64      `json:"id" form:"id"`                         //id
	MaxTries       int32       `json:"maxTries" form:"maxTries"`             //最大发送次数
	EmailAccountId *int64      `json:"emailAccountId" form:"emailAccountId"` //发件方
	To             string      `json:"to" form:"to"`                         //接收方
	IsSend         *bool       `json:"isSend" form:"isSend"`                 //是否已发送
	CreateTime     []time.Time `json:"createTime" form:"createTime"`         //时间
	request.Pagination
}

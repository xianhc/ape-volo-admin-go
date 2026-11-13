package queuedemail

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// QueuedEmailResponse 表示队列邮件接口的返回模型。
//
// 它显式列出邮件与审计字段，数据库实体不会进入 HTTP 输出。
type QueuedEmailResponse struct {
	Id                   int64       `json:"id,string"`
	From                 string      `json:"from"`
	FromName             string      `json:"fromName"`
	To                   string      `json:"to"`
	ToName               string      `json:"toName"`
	ReplyTo              string      `json:"replyTo"`
	ReplyToName          string      `json:"replyToName"`
	Priority             int32       `json:"priority"`
	Cc                   string      `json:"cc"`
	Bcc                  string      `json:"bcc"`
	Subject              string      `json:"subject"`
	Body                 string      `json:"body"`
	SentTries            int32       `json:"sentTries"`
	IsSend               bool        `json:"isSend"`
	SendTime             *model.Time `json:"sendTime"`
	EmailAccountId       int64       `json:"emailAccountId,string"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将队列邮件实体转换为接口返回模型。
//
// 转换保持现有 JSON 字段并隐藏软删除标记。
func fromEntity(email QueuedEmail) QueuedEmailResponse {
	return QueuedEmailResponse{Id: email.Id, From: email.From, FromName: email.FromName, To: email.To, ToName: email.ToName, ReplyTo: email.ReplyTo, ReplyToName: email.ReplyToName, Priority: email.Priority, Cc: email.Cc, Bcc: email.Bcc, Subject: email.Subject, Body: email.Body, SentTries: email.SentTries, IsSend: email.IsSend, SendTime: email.SendTime, EmailAccountId: email.EmailAccountId, CreatorId: email.CreatorId, CreationTime: email.CreationTime, LastModifierId: email.LastModifierId, LastModificationTime: email.LastModificationTime, DeleterId: email.DeleterId, DeletionTime: email.DeletionTime}
}

// fromEntities 将队列邮件实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(emails []QueuedEmail) []QueuedEmailResponse {
	result := make([]QueuedEmailResponse, 0, len(emails))
	for _, email := range emails {
		result = append(result, fromEntity(email))
	}
	return result
}

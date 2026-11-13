package queuedemail

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// QueuedEmail 表示持久化到数据库的待发送邮件。
//
// 实体保存收件信息、发送内容、重试次数与发送结果，不直接作为 HTTP 响应返回。
type QueuedEmail struct {
	model.RootKey
	From           string      `gorm:"comment:发件邮箱;not null;"`
	FromName       string      `gorm:"comment:发件人名称;"`
	To             string      `gorm:"comment:收件邮箱;not null;"`
	ToName         string      `gorm:"comment:收件人名称;"`
	ReplyTo        string      `gorm:"comment:回复邮箱;"`
	ReplyToName    string      `gorm:"comment:回复人名称;"`
	Priority       int32       `gorm:"comment:优先级;not null;"`
	Cc             string      `gorm:"comment:抄送;"`
	Bcc            string      `gorm:"comment:密件抄送;"`
	Subject        string      `gorm:"comment:标题;"`
	Body           string      `gorm:"comment:内容;not null"`
	SentTries      int32       `gorm:"index:index_queued_email_IsSend_SentTries,priority:2;comment:发送上限次数;not null;"`
	IsSend         bool        `gorm:"index:index_queued_email_IsSend_SentTries,priority:1;comment:是否已发送;not null;"`
	SendTime       *model.Time `gorm:"comment:发送时间;"`
	EmailAccountId int64       `gorm:"index:index_queued_email_EmailAccountId;comment:发件邮箱ID;not null;"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回队列邮件实体对应的数据库表名。
//
// GORM 在自动建表、查询和写入时统一使用该名称。
func (QueuedEmail) TableName() string { return "queued_email" }

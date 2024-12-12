package queued

import (
	"go-apevolo/model"
	"time"
)

type Email struct {
	model.RootKey
	To             string     `json:"to"  gorm:"comment:收件邮箱;not null;"`              // 收件邮箱
	ToName         string     `json:"toName" gorm:"comment:收件人名称;"`                   // 收件人名称
	ReplyTo        string     `json:"replyTo" gorm:"comment:回复邮箱;"`                   // 回复邮箱
	ReplyToName    string     `json:"replyToName"  gorm:"comment:回复人名称;"`             // 回复人名称
	Priority       int32      `json:"priority" gorm:"comment:优先级;not null;"`          // 优先级
	Cc             string     `json:"cc" gorm:"comment:抄送;"`                          // 抄送
	Bcc            string     `json:"bcc" gorm:"comment:密件抄送;"`                       // 密件抄送
	Subject        string     `json:"subject" gorm:"comment:标题;"`                     // 标题
	Body           string     `json:"body" gorm:"comment:内容;not null"`                // 内容
	SentTries      int32      `json:"sentTries" gorm:"comment:发送上限次数;not null;"`      // 发送上限次数
	SendTime       *time.Time `json:"sendTime" gorm:"comment:发送时间;"`                  // 发送时间
	EmailAccountId int64      `json:"emailAccountId" gorm:"comment:发件邮箱ID;not null;"` // 发件邮箱ID
	model.BaseModel
	model.SoftDeleted
}

func (Email) TableName() string {
	return "queued_email"
}

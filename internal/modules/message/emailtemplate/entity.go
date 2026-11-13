package emailtemplate

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// MessageTemplate 表示持久化到数据库的邮件模板。
//
// 实体保存模板正文与发送账户引用，不直接作为 HTTP 响应返回。
type MessageTemplate struct {
	model.RootKey
	Name              string                `gorm:"size:255;uniqueIndex:unique_email_message_template_Name;comment:模板名称;not null;"`
	BccEmailAddresses string                `gorm:"comment:抄送邮箱地址;"`
	Subject           string                `gorm:"comment:主题;"`
	Body              model.CustomFieldText `gorm:"type:varchar(4000);comment:内容;not null;"`
	Enabled           bool                  `gorm:"comment:是否激活;not null;"`
	EmailAccountId    int64                 `gorm:"index:index_email_message_template_EmailAccountId;comment:邮箱账户标识符;not null;"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回邮件模板实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (MessageTemplate) TableName() string { return "email_message_template" }

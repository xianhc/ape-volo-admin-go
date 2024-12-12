package email

import (
	"go-apevolo/model"
	"go-apevolo/utils"
)

type MessageTemplate struct {
	model.RootKey
	Name              string                `json:"name" gorm:"comment:模板名称;not null;"`                  // 模板名称
	BccEmailAddresses string                `json:"bccEmailAddresses" gorm:"comment:抄送邮箱地址;"`            // 抄送邮箱地址
	Subject           string                `json:"subject"  gorm:"comment:主题;"`                         // 主题
	Body              utils.CustomFieldText `json:"body" gorm:"type:varchar(4000);comment:内容;not null;"` // 内容
	IsActive          bool                  `json:"isActive" gorm:"comment:是否激活;not null;"`              // 是否激活
	EmailAccountId    int64                 `json:"emailAccountId"  gorm:"comment:邮箱账户标识符;not null;"`    // 邮箱账户标识符
	model.BaseModel
	model.SoftDeleted
}

func (MessageTemplate) TableName() string {
	return "email_message_template"
}

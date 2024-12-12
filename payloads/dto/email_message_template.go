package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateEmailMessageTemplateDto struct {
	model.RootKey
	Name              string `json:"name" validate:"required"`    // 模板名称
	BccEmailAddresses string `json:"bccEmailAddresses"`           // 抄送邮箱地址
	Subject           string `json:"subject" validate:"required"` // 主题
	Body              string `json:"body" validate:"required"`    // 内容
	IsActive          bool   `json:"isActive"`                    // 是否激活
	EmailAccountId    int64  `json:"emailAccountId"`              // 邮箱账户标识符
	model.BaseModel
}

func (req *CreateUpdateEmailMessageTemplateDto) Generate(model *email.MessageTemplate) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Name = req.Name
	model.BccEmailAddresses = req.BccEmailAddresses
	model.Subject = req.Subject
	model.Body = utils.CustomFieldText(req.Body)
	model.IsActive = req.IsActive
	model.EmailAccountId = req.EmailAccountId
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type EmailMessageTemplateQueryCriteria struct {
	Name       string      `json:"name" form:"name"`             //模板名称
	IsActive   *bool       `json:"isActive" form:"isActive"`     //是否激活
	CreateTime []time.Time `json:"createTime" form:"createTime"` //用户名
	request.Pagination
}

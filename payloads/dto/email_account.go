package dto

import (
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"time"
)

type CreateUpdateEmailAccountDto struct {
	model.RootKey
	Email                 string `json:"email" validate:"required,email"` // 电子邮件地址
	DisplayName           string `json:"displayName" validate:"required"` // 电子邮件显示名称
	Host                  string `json:"host" validate:"required"`        // 主机
	Port                  int32  `json:"port"`                            // 端口
	Username              string `json:"username" validate:"required"`    // 用户名
	Password              string `json:"password"  validate:"required"`   // 密码
	EnableSsl             bool   `json:"enableSsl"`                       // 是否SSL
	UseDefaultCredentials bool   `json:"useDefaultCredentials" `          // 是否与请求一起发送应用程序的默认系统凭据
	model.BaseModel
}

func (req *CreateUpdateEmailAccountDto) Generate(model *email.Account) {
	if req.Id != 0 {
		model.Id = req.Id
	} else {
		model.Id = int64(utils.GenerateID())
		model.CreateBy = req.CreateBy
		model.CreateTime = ext.GetCurrentTime()
	}
	model.Email = req.Email
	model.DisplayName = req.DisplayName
	model.Host = req.Host
	model.Port = req.Port
	model.Username = req.Username
	model.Password = req.Password
	model.EnableSsl = req.EnableSsl
	model.UseDefaultCredentials = req.UseDefaultCredentials
	if req.UpdateBy != nil {
		model.UpdateBy = req.UpdateBy
	}
	if req.UpdateTime != nil {
		localTime := ext.GetCurrentTime()
		model.UpdateTime = &localTime
	}
}

type EmailAccountQueryCriteria struct {
	Email       string      `json:"email" form:"email"`             //邮箱
	DisplayName string      `json:"displayName" form:"displayName"` //显示名称
	Username    string      `json:"username" form:"username"`       //
	CreateTime  []time.Time `json:"createTime" form:"createTime"`   //用户名
	request.Pagination
}

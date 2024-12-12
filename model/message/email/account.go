package email

import (
	"go-apevolo/model"
)

type Account struct {
	model.RootKey
	Email                 string `json:"email" gorm:"comment:电子邮件地址;not null;"`                               // 电子邮件地址
	DisplayName           string `json:"displayName" gorm:"comment:电子邮件显示名称;not null;"`                       // 电子邮件显示名称
	Host                  string `json:"host"  gorm:"comment:主机;not null;"`                                   // 主机
	Port                  int32  `json:"port" gorm:"comment:端口;not null;"`                                    // 端口
	Username              string `json:"username" gorm:"comment:用户名;not null;"`                               // 用户名
	Password              string `json:"password"  gorm:"comment:密码;not null;"`                               // 密码
	EnableSsl             bool   `json:"enableSsl" gorm:"comment:是否SSL;not null;"`                            // 是否SSL
	UseDefaultCredentials bool   `json:"useDefaultCredentials" gorm:"comment:是否与请求一起发送应用程序的默认系统凭据;not null;"` // 是否与请求一起发送应用程序的默认系统凭据
	model.BaseModel
	model.SoftDeleted
}

func (Account) TableName() string {
	return "email_account"
}

package emailaccount

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Account 表示持久化到数据库的邮件发送账户。
//
// Password 仅供内部发信流程使用，Entity 不直接作为 HTTP 响应返回。
type Account struct {
	model.RootKey
	Email                 string `gorm:"size:255;uniqueIndex:unique_email_account_Email;comment:电子邮件地址;not null;"`
	DisplayName           string `gorm:"comment:电子邮件显示名称;not null;"`
	Host                  string `gorm:"comment:主机;not null;"`
	Port                  int32  `gorm:"comment:端口;not null;"`
	UserName              string `gorm:"comment:用户名;not null;"`
	Password              string `gorm:"comment:密码;not null;"`
	EnableSsl             bool   `gorm:"comment:是否SSL;not null;"`
	UseDefaultCredentials bool   `gorm:"comment:是否与请求一起发送应用程序的默认系统凭据;not null;"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回邮件账户实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称。
func (Account) TableName() string { return "email_account" }

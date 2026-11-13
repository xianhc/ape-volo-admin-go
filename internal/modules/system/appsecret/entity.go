package appsecret

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// AppSecret 表示持久化到数据库的应用密钥。
//
// 该类型只描述 GORM 字段与审计能力，不直接作为 HTTP 响应返回。
type AppSecret struct {
	model.RootKey
	AppId        string `gorm:"comment:应用ID;type:varchar(100)"`
	AppSecretKey string `gorm:"comment:应用密钥"`
	AppName      string `gorm:"comment:应用名称"`
	Remark       string `gorm:"comment:描述"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回应用密钥实体对应的数据库表名。
//
// GORM 在自动建表、查询和写入时统一使用该名称。
func (AppSecret) TableName() string {
	return "sys_app_secret"
}

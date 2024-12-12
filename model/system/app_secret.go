package system

import (
	"go-apevolo/model"
)

type AppSecret struct {
	model.RootKey
	AppId        string `json:"appId" gorm:"comment:应用ID;type:varchar(100)"` // 应用ID
	AppSecretKey string `json:"appSecretKey" gorm:"comment:应用密钥"`            // 应用密钥
	AppName      string `json:"appName" gorm:"comment:应用名称"`                 // 应用名称
	Remark       string `json:"remark" gorm:"comment:描述"`                    // 描述
	model.BaseModel
	model.SoftDeleted
}

func (AppSecret) TableName() string {
	return "sys_app_secret"
}

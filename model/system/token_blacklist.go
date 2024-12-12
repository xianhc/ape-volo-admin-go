package system

import (
	"go-apevolo/model"
)

type TokenBlacklist struct {
	model.RootKey
	AccessToken string `json:"accessToken" gorm:"comment:令牌 登录token的MD5值"` // 令牌 登录token的MD5值
	model.BaseModel
	model.SoftDeleted
}

func (TokenBlacklist) TableName() string {
	return "sys_token_blacklist"
}

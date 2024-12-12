package system

import (
	"go-apevolo/model"
)

type Setting struct {
	model.RootKey
	Name        string `json:"name" gorm:"comment:设置键"`       // 设置键
	Value       string `json:"value" gorm:"comment:设置值"`      // 设置值
	Enabled     bool   `json:"enabled"  gorm:"comment:是否启用"`  // 是否启用
	Description string `json:"description" gorm:"comment:描述"` // 描述
	model.BaseModel
	model.SoftDeleted
}

func (Setting) TableName() string {
	return "sys_setting"
}

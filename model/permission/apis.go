package permission

import (
	"go-apevolo/model"
)

type Apis struct {
	model.RootKey
	Group       string `json:"group" gorm:"comment:组;not null;"`         // 组
	Url         string `json:"url" gorm:"comment:请求路径;not null;"`        // 请求路径
	Description string `json:"description"  gorm:"comment:描述;not null;"` // 描述
	Method      string `json:"method" gorm:"comment:请求方法;not null;"`     // 请求方法
	model.BaseModel
	model.SoftDeleted
}

func (Apis) TableName() string {
	return "sys_apis"
}

package permission

import (
	"go-apevolo/model"
)

type Job struct {
	model.RootKey
	Name    string `json:"name" gorm:"comment:岗位名称"`     // 岗位名称
	Sort    int    `json:"sort" gorm:"comment:排序"`       // 排序
	Enabled bool   `json:"enabled"  gorm:"comment:是否启用"` // 是否启用
	model.BaseModel
	model.SoftDeleted
}

func (Job) TableName() string {
	return "sys_job"
}

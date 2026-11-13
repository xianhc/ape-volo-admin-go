package setting

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// Setting 表示持久化到数据库的系统设置。
//
// 该类型只描述 GORM 映射和审计能力，不直接作为 HTTP 响应返回。
type Setting struct {
	model.RootKey
	Name        string `gorm:"size:255;uniqueIndex:unique_sys_setting_Name;comment:设置键"`
	Value       string `gorm:"comment:设置值"`
	Enabled     bool   `gorm:"comment:是否启用"`
	Description string `gorm:"comment:描述"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回系统设置实体对应的数据库表名。
//
// GORM 和种子初始化统一使用该名称定位数据表及资源文件。
func (Setting) TableName() string {
	return "sys_setting"
}

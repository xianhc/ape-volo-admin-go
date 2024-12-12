package utils

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// CustomFieldText 自定义字段文本类型
type CustomFieldText string

func (CustomFieldText) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// 根据数据库类型返回不同的字段类型
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "TEXT"
	case "sqlite":
		return "TEXT"
	case "mssql":
		return "NVARCHAR(MAX)"
	default:
		return "TEXT"
	}
}

package model

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// CustomFieldText 大文本列类型：按数据库方言选择对应的大字段类型。
//
// 放在 model 而不是 utils：它是持久化层类型，模型定义要用它；
// 留在工具包里会让 model 反向依赖 utils，把分层方向拧反。
type CustomFieldText string

func (text CustomFieldText) Value() (driver.Value, error) {
	return string(text), nil
}

func (text *CustomFieldText) Scan(value interface{}) error {
	switch value := value.(type) {
	case nil:
		*text = ""
	case string:
		*text = CustomFieldText(value)
	case []byte:
		*text = CustomFieldText(value)
	default:
		return fmt.Errorf("unsupported text value %T", value)
	}
	return nil
}

// GormDBDataType 返回当前数据库方言下的大文本类型。
func (CustomFieldText) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "TEXT"
	case "sqlite":
		return "TEXT"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	case "oracle":
		return "CLOB"
	default:
		return "TEXT"
	}
}

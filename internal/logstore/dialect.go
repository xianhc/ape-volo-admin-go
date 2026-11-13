package logstore

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

// Quote 按日志连接方言引用受控标识符。
//
// 调用方仍须先验证表名与列名，不能传入请求原文。
func Quote(db *gorm.DB, name string) string {
	stmt := &gorm.Statement{DB: db}
	return stmt.Quote(name)
}

// DayExpression 对已经按配置时区写入的 creation_time 分组。
//
// 带时区的列先转换成日志配置时区的日期。
func DayExpression(db *gorm.DB, zone string) (string, error) {
	c := Quote(db, "creation_time")
	switch db.Dialector.Name() {
	case "sqlite":
		return "substr(" + c + ", 1, 10)", nil
	case "mysql":
		return "DATE_FORMAT(" + c + ", '%Y-%m-%d')", nil
	case "postgres":
		return "to_char(" + c + " AT TIME ZONE '" + strings.ReplaceAll(zone, "'", "''") + "', 'YYYY-MM-DD')", nil
	case "sqlserver":
		return "CONVERT(varchar(10), " + c + ", 23)", nil
	case "oracle":
		return "TO_CHAR(" + c + ", 'YYYY-MM-DD')", nil
	default:
		return "", fmt.Errorf("不支持日志日期方言")
	}
}

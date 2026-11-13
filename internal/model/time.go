package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TimeLayout 是时间字段对外的统一格式。
//
// 前端表格直接渲染 `creationTime` 等字段，不额外格式化；
// 用 time.Time 默认的 RFC3339（2026-11-03T10:58:37.48+08:00）会显示成难读的 ISO 串。
const TimeLayout = "2006-01-02 15:04:05"

// Time 是实体中所有时间字段的类型，对外统一按 TimeLayout 序列化。
//
// 内嵌 time.Time 以保留 Format / After / Before / IsZero 等方法，
// 同时实现 driver.Valuer 与 sql.Scanner，让 GORM 仍按单个时间列读写。
type Time struct {
	time.Time
}

// Now 返回当前时间。
func Now() Time {
	return Time{Time: time.Now()}
}

// MarshalJSON 输出 `yyyy-MM-dd HH:mm:ss`；零值输出 null，避免出现 0001-01-01。
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format(TimeLayout) + `"`), nil
}

// UnmarshalJSON 同时接受 `yyyy-MM-dd HH:mm:ss` 与 RFC3339 两种写法。
//
// 前者是本项目对外的格式，后者是前端可能原样回传的浏览器标准格式，
// 只认一种会让「查出来再存回去」的请求失败。
func (t *Time) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	parsed, err := parseTime(s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Value 实现 driver.Valuer。
func (t Time) Value() (driver.Value, error) {
	return t.Time, nil
}

// Scan 实现 sql.Scanner，兼容驱动返回的 time.Time / string / []byte。
func (t *Time) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		t.Time = time.Time{}
		return nil
	case time.Time:
		t.Time = v
		return nil
	case string:
		parsed, err := parseTime(v)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case []byte:
		parsed, err := parseTime(string(v))
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("不支持的时间类型 %T", value)
	}
}

// GormDataType 声明列类型，避免 GORM 把内嵌结构体当成组合字段展开。
func (Time) GormDataType() string {
	return "datetime"
}

// GormDBDataType 为时间字段选择当前连接可接受的列类型。
//
// 带时区的数据库保留时间点，日志统计再按配置时区转换。
func (Time) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "TIMESTAMPTZ"
	case "sqlserver":
		return "DATETIMEOFFSET"
	case "oracle":
		return "TIMESTAMP WITH TIME ZONE"
	case "mysql":
		return "DATETIME(6)"
	default:
		return "DATETIME"
	}
}

// parseTime 解析项目格式与 RFC3339。
func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{TimeLayout, time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间 %q", s)
}

// Package logstore 管理日志连接上的受控月表及跨月读取。
package logstore

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// 月表基名仅由此处受控常量提供。
//
// 实际表名追加月份首日后缀，不接收请求指定的表名。
const (
	Operate   = "log_operate"
	Exception = "log_exception"
	Job       = "log_job"
)

var current atomic.Pointer[Store]

// SetDefault 发布已初始化的日志存储。
//
// 启动时设置，消费者结束前不替换。
func SetDefault(s *Store) { current.Store(s) }

// Default 返回当前进程的日志存储。
//
// 未初始化时返回 nil，由仓储报告不可用。
func Default() *Store { return current.Load() }

func validPrefix(prefix string) bool {
	return prefix == Operate || prefix == Exception || prefix == Job
}

// MonthTable 按指定时区生成月份首日表名。
//
// 类型必须属于三种受控基名，拒绝零时间。
func MonthTable(prefix string, at time.Time, loc *time.Location) (string, error) {
	if !validPrefix(prefix) || at.IsZero() || loc == nil {
		return "", fmt.Errorf("无效的日志月份或类型")
	}
	return prefix + "_" + at.In(loc).Format("200601") + "01", nil
}

func tableMonth(prefix, table string, loc *time.Location) (time.Time, bool) {
	if !validPrefix(prefix) || !strings.HasPrefix(table, prefix+"_") {
		return time.Time{}, false
	}
	suffix := strings.TrimPrefix(table, prefix+"_")
	if len(suffix) != 8 || !strings.HasSuffix(suffix, "01") {
		return time.Time{}, false
	}
	at, err := time.ParseInLocation("20060102", suffix, loc)
	return at, err == nil && at.Format("20060102") == suffix
}

// Location 返回已装配存储的日志时区。
//
// 生产入口在存储尚未装配时仍可生成时间快照。
func Location() *time.Location {
	if s := Default(); s != nil {
		return s.Location
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		return time.UTC
	}
	return loc
}

// Now 生成按日志时区解释的生产时刻。
//
// 统一到微秒精度，避免数据库精度舍入把月末记录移到次月时刻。
func Now() time.Time { return time.Now().In(Location()).Truncate(time.Microsecond) }

package ext

import (
	"go-apevolo/global"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func StringToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		return 0
	}
	return val
}

func StringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		return 0
	}
	return val
}

func GetTimestamp(date time.Time) (timestamp int64) {
	return date.UnixNano() / int64(time.Millisecond)
}

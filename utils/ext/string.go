package ext

import (
	"encoding/json"
	"strconv"
	"strings"
)

func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}

// GetCurrentTimeStr 获取当前时间字符串
func GetCurrentTimeStr() string {
	return GetCurrentTime().Format("2006-01-02 15:04:05")
}

//StructToJsonStr 结构体转json
func StructToJsonStr(e interface{}) (string, error) {
	if b, err := json.Marshal(e); err == nil {
		return string(b), err
	} else {
		return "", err
	}
}

// StringReplace 字符串替换
func StringReplace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

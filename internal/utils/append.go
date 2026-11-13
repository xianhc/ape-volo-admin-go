package utils

// AppendIfNotExists 追加字符串，已存在则原样返回。
// 空串视为无效值，直接跳过。
func AppendIfNotExists(slice []string, item string) []string {
	if item == "" {
		return slice
	}
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}

// AppendInt64 追加 int64，已存在则原样返回。
// 0 视为无效值，直接跳过。
func AppendInt64(slice []int64, item int64) []int64 {
	if item == 0 {
		return slice
	}
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}

// ContainsValue 判断字符串切片是否包含指定值
func ContainsValue(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

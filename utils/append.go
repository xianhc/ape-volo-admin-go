package utils

import (
	"github.com/emirpasic/gods/sets/hashset"
)

func AppendIfNotExists(slice []string, item string) []string {
	if item == "" {
		return slice
	}
	set := hashset.New()

	for _, val := range slice {
		set.Add(val)
	}

	// 如果元素已存在，则不添加
	if set.Contains(item) {
		return slice
	}

	// 元素不存在，添加到切片末尾
	return append(slice, item)
}

func AppendInt64(slice []int64, item int64) []int64 {
	if item == 0 {
		return slice
	}
	set := hashset.New()

	for _, val := range slice {
		set.Add(val)
	}

	// 如果元素已存在，则不添加
	if set.Contains(item) {
		return slice
	}

	// 元素不存在，添加到切片末尾
	return append(slice, item)
}

func ContainsValue(slice []string, value string) bool {
	valueMap := make(map[string]bool)
	for _, item := range slice {
		valueMap[item] = true
	}
	return valueMap[value]
}

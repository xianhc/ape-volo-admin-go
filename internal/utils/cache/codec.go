package cache

import (
	"encoding/json"
	"fmt"
)

// encode 把任意值序列化为存储用的字节。
//
// 使用标准库 encoding/json（不引入 sonic 等替换实现），原因：
//   - 缓存层是低频路径，序列化开销远小于网络/锁竞争；
//   - 避免与 ext 包形成循环依赖（ext -> global -> cache）。
func encode(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("cache: encode value: %w", err)
	}
	return b, nil
}

// decode 把存储的字节反序列化到 dest。
func decode(data []byte, dest any) error {
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache: decode value: %w", err)
	}
	return nil
}

// decodeInt 把存储的字节解析为整数，供 Incr 使用。
//
// 兼容两种写入方式：JSON 数字（Set(1) 得到 "1"）与纯十进制字符串。
func decodeInt(data []byte) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return 0, fmt.Errorf("cache: value is not an integer: %w", err)
	}
	return n, nil
}

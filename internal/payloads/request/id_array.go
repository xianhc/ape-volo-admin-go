package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// IDArray 是一组主键，反序列化时同时接受数字与字符串两种写法。
//
// 背景：本项目的响应体把 ID 序列化成字符串（`json:"id,string"`），
// 前端很可能把拿到的字符串原样回传，若只接受数字会直接 400。
// 这里两种都收，避免前后端对不齐。
type IDArray []int64

// UnmarshalJSON 实现 json.Unmarshaler。
func (a *IDArray) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*a = nil
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("IDArray 需要数组，实际: %s", string(data))
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	ids := make([]int64, 0, len(raw))
	for _, item := range raw {
		id, err := parseID(item)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	*a = ids
	return nil
}

// parseID 解析单个 ID，兼容 123 / "123" / null。
func parseID(raw json.RawMessage) (int64, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return 0, err
		}
		if s == "" {
			return 0, nil
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("无法解析 ID %q: %w", s, err)
		}
		return id, nil
	}

	var id int64
	if err := json.Unmarshal(raw, &id); err != nil {
		return 0, fmt.Errorf("无法解析 ID %s: %w", string(raw), err)
	}
	return id, nil
}

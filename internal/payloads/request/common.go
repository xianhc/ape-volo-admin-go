package request

import (
	"bytes"
	"encoding/json"
	"errors"
)

type IdCollection struct {
	IdArray []string `json:"idArray" form:"idArray"` //id集合
}

// UnmarshalJSON 兼容三种主键集合写法：
//
//	{"ids": ["1","2"]}、{"idArray": [1,2]}、裸数组 [1,2]
//
// 前两种是对象包裹形式，裸数组是不带字段名的写法。三种都要认：
// 前端不同版本、不同调用点传法不一致，只认一种会直接 400。
// 数字一律转成字符串保存，避免 int64 主键在 float64 解析中丢精度。
func (c *IdCollection) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	raw := data
	if data[0] == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		raw = nil
		for _, key := range []string{"ids", "idArray"} {
			if v, ok := obj[key]; ok && len(v) > 0 && string(v) != "null" {
				raw = v
				break
			}
		}
		if raw == nil {
			return errors.New("缺少 ids 字段")
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var items []any
	if err := decoder.Decode(&items); err != nil {
		return err
	}

	ids := make([]string, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case string:
			ids = append(ids, v)
		case json.Number:
			ids = append(ids, v.String())
		default:
			return errors.New("ids 只能是数字或字符串数组")
		}
	}
	c.IdArray = ids
	return nil
}

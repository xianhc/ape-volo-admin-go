package response

// ActionResultPage 分页结果。
//
// 字段为 `items` / `totalCount`，前端分页组件直接消费。
type ActionResultPage struct {
	Items      interface{} `json:"items"`      // 当前页结果项
	TotalCount int64       `json:"totalCount"` // 分页前的总条数
}

// ActionResult 操作结果。
//
// `isSuccess` 供前端统一判断成功与否；其余字段保留，便于排障与兼容现有前端。
type ActionResult struct {
	IsSuccess   bool         `json:"isSuccess"`   // 是否成功
	Status      int          `json:"status"`      // 状态码
	ActionError *ActionError `json:"actionError"` // 错误集合
	Message     string       `json:"message"`     // 消息提示
	Timestamp   int64        `json:"timestamp"`   // 时间戳
	Path        string       `json:"path"`        // 请求路径
}

type ActionError struct {
	Errors map[string]string `json:"errors,omitempty"`
}

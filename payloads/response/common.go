package response

type ActionResultPage struct {
	Content       interface{} `json:"content"`       //响应内容
	TotalElements int64       `json:"totalElements"` //总条数
}

type ActionResult struct {
	Status      int          `json:"status"`      //状态码
	ActionError *ActionError `json:"actionError"` //错误集合
	Message     string       `json:"message"`     //消息提示
	Timestamp   int64        `json:"timestamp"`   //时间戳
	Path        string       `json:"path"`        //请求路径
}

type ActionError struct {
	Errors map[string]string `json:"errors,omitempty"`
}

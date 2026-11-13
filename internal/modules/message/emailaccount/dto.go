package emailaccount

// Credentials 表示内部发信流程使用的 SMTP 凭据。
//
// 该 DTO 不进入 HTTP 响应，仅在 Service、任务和队列业务之间传递。
type Credentials struct {
	Id          int64
	Email       string
	DisplayName string
	Password    string
	Host        string
	Port        int32
	EnableSSL   bool
}

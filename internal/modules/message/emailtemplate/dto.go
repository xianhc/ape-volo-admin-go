package emailtemplate

// TemplateContent 表示内部邮件组装流程使用的模板内容。
//
// 该 DTO 不进入 HTTP 响应，只在模板 Service 与队列业务之间传递。
type TemplateContent struct {
	BccEmailAddresses string
	Subject           string
	Body              string
	EmailAccountId    int64
}

package emailtemplate

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// TemplateResponse 表示邮件模板接口的返回模型。
//
// 它显式列出模板与审计字段，数据库实体不会进入 HTTP 输出。
type TemplateResponse struct {
	Id                   int64                 `json:"id,string"`
	Name                 string                `json:"name"`
	BccEmailAddresses    string                `json:"bccEmailAddresses"`
	Subject              string                `json:"subject"`
	Body                 model.CustomFieldText `json:"body"`
	Enabled              bool                  `json:"enabled"`
	EmailAccountId       int64                 `json:"emailAccountId,string"`
	CreatorId            *int64                `json:"creatorId,omitempty,string"`
	CreationTime         model.Time            `json:"creationTime"`
	LastModifierId       *int64                `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time           `json:"lastModificationTime,omitempty"`
	DeleterId            *int64                `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time           `json:"deletionTime,omitempty"`
}

// fromEntity 将邮件模板实体转换为接口返回模型。
//
// 转换保持现有字段与正文类型不变。
func fromEntity(template MessageTemplate) TemplateResponse {
	return TemplateResponse{Id: template.Id, Name: template.Name, BccEmailAddresses: template.BccEmailAddresses, Subject: template.Subject, Body: template.Body, Enabled: template.Enabled, EmailAccountId: template.EmailAccountId, CreatorId: template.CreatorId, CreationTime: template.CreationTime, LastModifierId: template.LastModifierId, LastModificationTime: template.LastModificationTime, DeleterId: template.DeleterId, DeletionTime: template.DeletionTime}
}

// fromEntities 将邮件模板实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(templates []MessageTemplate) []TemplateResponse {
	result := make([]TemplateResponse, 0, len(templates))
	for _, template := range templates {
		result = append(result, fromEntity(template))
	}
	return result
}

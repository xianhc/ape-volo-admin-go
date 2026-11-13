package emailaccount

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// AccountResponse 表示邮件账户接口的返回模型。
//
// Password 始终为空，接口只返回账户配置而不泄露发信凭据。
type AccountResponse struct {
	Id                    int64       `json:"id,string"`
	Email                 string      `json:"email"`
	DisplayName           string      `json:"displayName"`
	Host                  string      `json:"host"`
	Port                  int32       `json:"port"`
	UserName              string      `json:"userName"`
	Password              string      `json:"password"`
	EnableSsl             bool        `json:"enableSsl"`
	UseDefaultCredentials bool        `json:"useDefaultCredentials"`
	CreatorId             *int64      `json:"creatorId,omitempty,string"`
	CreationTime          model.Time  `json:"creationTime"`
	LastModifierId        *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime  *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId             *int64      `json:"deleterId,omitempty,string"`
	DeletionTime          *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将邮件账户实体转换为脱敏接口返回模型。
//
// 转换有意不复制 Password 字段。
func fromEntity(account Account) AccountResponse {
	return AccountResponse{Id: account.Id, Email: account.Email, DisplayName: account.DisplayName, Host: account.Host, Port: account.Port, UserName: account.UserName, EnableSsl: account.EnableSsl, UseDefaultCredentials: account.UseDefaultCredentials, CreatorId: account.CreatorId, CreationTime: account.CreationTime, LastModifierId: account.LastModifierId, LastModificationTime: account.LastModificationTime, DeleterId: account.DeleterId, DeletionTime: account.DeletionTime}
}

// fromEntities 将邮件账户实体列表转换为脱敏返回模型列表。
//
// 返回值始终是非 nil 切片。
func fromEntities(accounts []Account) []AccountResponse {
	result := make([]AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, fromEntity(account))
	}
	return result
}

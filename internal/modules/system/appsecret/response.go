package appsecret

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// AppSecretResponse 表示应用密钥接口的返回模型。
//
// 它显式列出可公开字段，使数据库实体与 HTTP 契约保持解耦。
type AppSecretResponse struct {
	Id                   int64       `json:"id,string"`
	AppId                string      `json:"appId"`
	AppSecretKey         string      `json:"appSecretKey"`
	AppName              string      `json:"appName"`
	Remark               string      `json:"remark"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将应用密钥实体转换为接口返回模型。
//
// 转换只复制公开字段，不把 GORM 元数据或软删除标记暴露给客户端。
func fromEntity(appSecret AppSecret) AppSecretResponse {
	return AppSecretResponse{
		Id:                   appSecret.Id,
		AppId:                appSecret.AppId,
		AppSecretKey:         appSecret.AppSecretKey,
		AppName:              appSecret.AppName,
		Remark:               appSecret.Remark,
		CreatorId:            appSecret.CreatorId,
		CreationTime:         appSecret.CreationTime,
		LastModifierId:       appSecret.LastModifierId,
		LastModificationTime: appSecret.LastModificationTime,
		DeleterId:            appSecret.DeleterId,
		DeletionTime:         appSecret.DeletionTime,
	}
}

// fromEntities 将应用密钥实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片，空查询会序列化为 JSON 空数组。
func fromEntities(appSecrets []AppSecret) []AppSecretResponse {
	result := make([]AppSecretResponse, 0, len(appSecrets))
	for _, appSecret := range appSecrets {
		result = append(result, fromEntity(appSecret))
	}
	return result
}

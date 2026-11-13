package setting

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// SettingResponse 表示系统设置接口的返回模型。
//
// 它显式列出公开字段，使数据库实体与 HTTP 契约保持解耦。
type SettingResponse struct {
	Id                   int64       `json:"id,string"`
	Name                 string      `json:"name"`
	Value                string      `json:"value"`
	Enabled              bool        `json:"enabled"`
	Description          string      `json:"description"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将系统设置实体转换为接口返回模型。
//
// 转换不暴露 GORM 元数据和软删除标记。
func fromEntity(setting Setting) SettingResponse {
	return SettingResponse{
		Id:                   setting.Id,
		Name:                 setting.Name,
		Value:                setting.Value,
		Enabled:              setting.Enabled,
		Description:          setting.Description,
		CreatorId:            setting.CreatorId,
		CreationTime:         setting.CreationTime,
		LastModifierId:       setting.LastModifierId,
		LastModificationTime: setting.LastModificationTime,
		DeleterId:            setting.DeleterId,
		DeletionTime:         setting.DeletionTime,
	}
}

// fromEntities 将系统设置实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片，空查询会序列化为 JSON 空数组。
func fromEntities(settings []Setting) []SettingResponse {
	result := make([]SettingResponse, 0, len(settings))
	for _, setting := range settings {
		result = append(result, fromEntity(setting))
	}
	return result
}

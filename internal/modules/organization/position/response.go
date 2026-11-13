package position

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// PositionResponse 表示岗位接口的返回模型。
//
// 它显式列出岗位与审计字段，数据库实体不会进入 HTTP 输出。
type PositionResponse struct {
	Id                   int64       `json:"id,string"`
	Name                 string      `json:"name"`
	Sort                 int         `json:"sort"`
	Enabled              bool        `json:"enabled"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将岗位实体转换为接口返回模型。
//
// 转换保持现有 JSON 字段并隐藏软删除标记。
func fromEntity(position Position) PositionResponse {
	return PositionResponse{Id: position.Id, Name: position.Name, Sort: position.Sort, Enabled: position.Enabled, CreatorId: position.CreatorId, CreationTime: position.CreationTime, LastModifierId: position.LastModifierId, LastModificationTime: position.LastModificationTime, DeleterId: position.DeleterId, DeletionTime: position.DeletionTime}
}

// FromEntities 将岗位实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func FromEntities(positions []Position) []PositionResponse {
	result := make([]PositionResponse, 0, len(positions))
	for _, position := range positions {
		result = append(result, fromEntity(position))
	}
	return result
}

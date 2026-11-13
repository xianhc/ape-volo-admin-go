package dict

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// DictDetailResponse 表示字典明细接口的返回模型。
//
// 所属字典由 DictId 表示，不重复嵌入父对象。
type DictDetailResponse struct {
	Id                   int64       `json:"id,string"`
	DictId               int64       `json:"dictId,string"`
	Label                string      `json:"label"`
	Value                string      `json:"value"`
	DictSort             int         `json:"dictSort"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// DictResponse 表示字典主表接口的返回模型。
//
// DictDetails 使用独立明细 Response，数据库模型不会进入 HTTP 输出。
type DictResponse struct {
	Id                   int64                `json:"id,string"`
	DictType             int                  `json:"dictType"`
	Name                 string               `json:"name"`
	Description          string               `json:"description"`
	CreatorId            *int64               `json:"creatorId,omitempty,string"`
	CreationTime         model.Time           `json:"creationTime"`
	LastModifierId       *int64               `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time          `json:"lastModificationTime,omitempty"`
	DeleterId            *int64               `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time          `json:"deletionTime,omitempty"`
	DictDetails          []DictDetailResponse `json:"dictDetails"`
}

// detailFromEntity 将字典明细实体转换为接口返回模型。
//
// 所属字典只保留主键，避免在主从嵌套响应中重复输出父对象。
func detailFromEntity(detail DictDetail) DictDetailResponse {
	return DictDetailResponse{Id: detail.Id, DictId: detail.DictId, Label: detail.Label, Value: detail.Value, DictSort: detail.DictSort, CreatorId: detail.CreatorId, CreationTime: detail.CreationTime, LastModifierId: detail.LastModifierId, LastModificationTime: detail.LastModificationTime, DeleterId: detail.DeleterId, DeletionTime: detail.DeletionTime}
}

// detailsFromEntities 将字典明细实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片，空字典会序列化为空数组。
func detailsFromEntities(details []DictDetail) []DictDetailResponse {
	result := make([]DictDetailResponse, 0, len(details))
	for _, detail := range details {
		result = append(result, detailFromEntity(detail))
	}
	return result
}

// fromEntity 将字典实体及预加载明细转换为接口返回模型。
//
// 转换保留主从嵌套 JSON 契约并隐藏软删除标记。
func fromEntity(dict Dict) DictResponse {
	return DictResponse{Id: dict.Id, DictType: dict.DictType, Name: dict.Name, Description: dict.Description, CreatorId: dict.CreatorId, CreationTime: dict.CreationTime, LastModifierId: dict.LastModifierId, LastModificationTime: dict.LastModificationTime, DeleterId: dict.DeleterId, DeletionTime: dict.DeletionTime, DictDetails: detailsFromEntities(dict.DictDetails)}
}

// fromEntities 将字典实体列表转换为接口返回模型列表。
//
// 每个字典的预加载明细一并完成转换。
func fromEntities(dicts []Dict) []DictResponse {
	result := make([]DictResponse, 0, len(dicts))
	for _, dict := range dicts {
		result = append(result, fromEntity(dict))
	}
	return result
}

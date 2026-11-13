package filerecord

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// FileRecordResponse 表示文件记录接口的返回模型。
//
// 它显式列出文件元数据和审计字段，不暴露持久化实现细节。
type FileRecordResponse struct {
	Id                   int64       `json:"id,string"`
	Description          string      `json:"description"`
	ContentType          string      `json:"contentType"`
	ContentTypeName      string      `json:"contentTypeName"`
	ContentTypeNameEn    string      `json:"contentTypeNameEn"`
	OriginalName         string      `json:"originalName"`
	NewName              string      `json:"newName"`
	FilePath             string      `json:"filePath"`
	Size                 string      `json:"size"`
	CreatorId            *int64      `json:"creatorId,omitempty,string"`
	CreationTime         model.Time  `json:"creationTime"`
	LastModifierId       *int64      `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time `json:"lastModificationTime,omitempty"`
	DeleterId            *int64      `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time `json:"deletionTime,omitempty"`
}

// fromEntity 将文件记录实体转换为接口返回模型。
//
// 转换保留现有 JSON 字段，同时隐藏软删除标记。
func fromEntity(file FileRecord) FileRecordResponse {
	return FileRecordResponse{
		Id: file.Id, Description: file.Description, ContentType: file.ContentType,
		ContentTypeName: file.ContentTypeName, ContentTypeNameEn: file.ContentTypeNameEn,
		OriginalName: file.OriginalName, NewName: file.NewName, FilePath: file.FilePath, Size: file.Size,
		CreatorId: file.CreatorId, CreationTime: file.CreationTime, LastModifierId: file.LastModifierId,
		LastModificationTime: file.LastModificationTime, DeleterId: file.DeleterId, DeletionTime: file.DeletionTime,
	}
}

// fromEntities 将文件记录实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片，空查询会序列化为 JSON 空数组。
func fromEntities(files []FileRecord) []FileRecordResponse {
	result := make([]FileRecordResponse, 0, len(files))
	for _, file := range files {
		result = append(result, fromEntity(file))
	}
	return result
}

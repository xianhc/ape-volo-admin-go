package system

import (
	"go-apevolo/model"
)

type FileRecord struct {
	model.RootKey
	Description       string `json:"description" gorm:"comment:文件描述"`           // 文件描述
	ContentType       string `json:"contentType" gorm:"comment:文件类型"`           // 文件类型
	ContentTypeName   string `json:"contentTypeName" gorm:"comment:文件类别"`       // 文件类别
	ContentTypeNameEn string `json:"contentTypeNameEn" gorm:"comment:文件类别英文名称"` // 文件类别英文名称
	OriginalName      string `json:"originalName" gorm:"comment:文件原名称"`         // 文件原名称
	NewName           string `json:"newName" gorm:"comment:文件新名称"`              // 文件新名称
	FilePath          string `json:"filePath" gorm:"comment:文件存储路径"`            // 文件存储路径
	Size              string `json:"size" gorm:"comment:文件大小"`                  // 文件大小
	model.BaseModel
	model.SoftDeleted
}

func (FileRecord) TableName() string {
	return "sys_file_record"
}

package filerecord

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// FileRecord 表示持久化到数据库的文件记录。
//
// 实体保存文件元数据和存储路径，不直接作为 HTTP 响应返回。
type FileRecord struct {
	model.RootKey
	Description       string `gorm:"comment:文件描述"`
	ContentType       string `gorm:"comment:文件类型"`
	ContentTypeName   string `gorm:"comment:文件类别"`
	ContentTypeNameEn string `gorm:"comment:文件类别英文名称"`
	OriginalName      string `gorm:"comment:文件原名称"`
	NewName           string `gorm:"comment:文件新名称"`
	FilePath          string `gorm:"comment:文件存储路径"`
	Size              string `gorm:"comment:文件大小"`
	model.BaseModel
	model.SoftDeleted
	model.DataScoped
}

// TableName 返回文件记录实体对应的数据库表名。
//
// GORM 在自动建表、查询和写入时统一使用该名称。
func (FileRecord) TableName() string {
	return "sys_file_record"
}

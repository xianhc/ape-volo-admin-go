package upload

import (
	"mime/multipart"

	"go-apevolo/global"
)

// OSS 对象存储接口
type OSS interface {
	UploadFile(file *multipart.FileHeader) (string, string, string, string, string, error)
	DeleteFile(key string) error
}

// NewOss OSS的实例化方法
func NewOss() OSS {
	switch global.Config.System.OssType {
	case "local":
		return &Local{}
	default:
		return &Local{}
	}
}

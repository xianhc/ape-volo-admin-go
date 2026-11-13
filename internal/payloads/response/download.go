package response

import (
	"mime"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// DownloadFile 以附件形式发送已生成的导出文件。
//
// 文件类型根据下载名称的扩展名确定，未知类型使用二进制流。
func DownloadFile(c *gin.Context, filePath, fileName string) {
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.FileAttachment(filePath, fileName)
}

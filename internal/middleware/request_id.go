package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// RequestID 为每个请求分配唯一标识，并回写到响应头 `X-Request-ID`。
//
// 客户端可自带该头（便于把前端错误与后端日志关联），但值会被校验与截断：
// 头部直接回写响应，不做限制等于让调用方注入任意内容。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := sanitizeRequestID(c.GetHeader("X-Request-ID"))
		if id == "" {
			id = utils.GenerateID().String()
		}
		utils.SetRequestID(c, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// sanitizeRequestID 只接受由字母、数字与 `.` `_` `-` `:` 组成的短字符串。
//
// 不合法时返回空串，由调用方重新生成，避免把客户端任意内容写进响应头。
func sanitizeRequestID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 64 {
		return ""
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.', r == '_', r == '-', r == ':':
		default:
			return ""
		}
	}
	return raw
}

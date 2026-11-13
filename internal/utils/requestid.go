package utils

import "github.com/gin-gonic/gin"

// requestIDKey 是写入 gin.Context 的键。
//
// 读写都收在 utils：中间件负责生成与回写响应头，接口层（如客户端错误上报）
// 需要读同一个键，键定义在中间件里会让接口层反向依赖中间件。
const requestIDKey = "requestId"

// SetRequestID 把请求标识写入上下文。
func SetRequestID(c *gin.Context, id string) {
	if c == nil {
		return
	}
	c.Set(requestIDKey, id)
}

// RequestIDFromContext 取当前请求的标识，没有则返回空串。
func RequestIDFromContext(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(requestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

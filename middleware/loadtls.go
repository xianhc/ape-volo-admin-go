package middleware

import (
	"go-apevolo/global"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func LoadTls() gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:443",
		})
		err := middleware.Process(c.Writer, c.Request)
		if err != nil {
			// 如果出现错误，请不要继续
			global.Logger.Error("加载tls失败", zap.Error(err))
			return
		}
		c.Next()
	}
}

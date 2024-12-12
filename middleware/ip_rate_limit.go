package middleware

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"time"
)

func IpLimit() gin.HandlerFunc {
	// 创建限流器
	captchaLimiter := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Second * 5})  // 5秒3次
	defaultLimiter := tollbooth.NewLimiter(8, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Second * 10}) // 默认10秒8次

	// 使用映射存储每个接口的限流器
	limiterMap := map[string]*limiter.Limiter{
		"/auth/captcha": captchaLimiter,
	}

	// 返回限流中间件函数
	return func(c *gin.Context) {
		// 获取请求路径
		path := c.FullPath()

		// 获取相应路径的限流器
		lim, exists := limiterMap[path]
		if !exists {
			// 如果路径不在映射中，则使用默认限流器
			lim = defaultLimiter
		}

		// 使用 IP 地址作为限流键
		httpError := tollbooth.LimitByKeys(lim, []string{c.ClientIP()})
		if httpError != nil {
			c.JSON(httpError.StatusCode, gin.H{
				"status":  httpError.StatusCode,
				"message": "访问过于频繁，请稍后重试！",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

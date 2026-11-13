package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ratelimit"
	"go.uber.org/zap"
)

// rateRule 描述一条路径的限流规则。
type rateRule struct {
	limit  int64
	window time.Duration
}

// defaultRateRule 是未命中任何规则的兜底限流。
var defaultRateRule = rateRule{limit: 8, window: 10 * time.Second}

// rateRules 按「请求路径」配置限流规则。
//
// 匹配使用 c.Request.URL.Path 而非 c.FullPath()：
// FullPath() 在路由未命中（404）时返回空串，会导致所有 404 请求
// 共用同一个限流桶，既放过了真实攻击路径，也误伤正常探测。
//
// 键是**完整路径**（含 system.router-prefix），由 rateRulesWithPrefix 拼装，
// 避免改了前缀之后登录接口悄悄退回兜底限流。
func rateRulesWithPrefix(prefix string) map[string]rateRule {
	p := "/" + strings.Trim(prefix, "/")
	if p == "/" {
		p = ""
	}
	return map[string]rateRule{
		p + "/auth/captcha":          {limit: 3, window: 5 * time.Second},
		p + "/auth/login":            {limit: 5, window: time.Minute},
		p + "/auth/reset-email-code": {limit: 1, window: time.Minute},
	}
}

// IpLimit 基于客户端 IP 的限流中间件。
//
// 限流后端跟随 system.use-redis：开启时走 Redis（多实例共享计数），
// 关闭时走进程内存（单实例）。
func IpLimit() (gin.HandlerFunc, func(), error) {
	rules := rateRulesWithPrefix(global.Config.System.RouterPrefix)
	limiters := make(map[string]ratelimit.Limiter, len(rules))
	var fallback ratelimit.Limiter
	closeLimiters := func() {
		for _, limiter := range limiters {
			_ = limiter.Close()
		}
		if fallback != nil {
			_ = fallback.Close()
		}
	}
	for path, rule := range rules {
		lim, err := newLimiter(rule)
		if err != nil {
			closeLimiters()
			return nil, nil, fmt.Errorf("初始化路径 %s 限流失败: %w", path, err)
		}
		limiters[path] = lim
	}
	var err error
	fallback, err = newLimiter(defaultRateRule)
	if err != nil {
		closeLimiters()
		return nil, nil, fmt.Errorf("初始化默认限流失败: %w", err)
	}
	return ipLimitHandler(limiters, fallback), closeLimiters, nil
}

func ipLimitHandler(limiters map[string]ratelimit.Limiter, fallback ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		lim, ok := limiters[path]
		if !ok {
			lim = fallback
		}

		// 限流键包含路径，避免登录接口的配额被其他接口消耗
		key := path + ":" + utils.GetClientIP(c)
		result, err := lim.Allow(c.Request.Context(), key)
		if err != nil {
			// 限流组件故障时放行而不是拒绝：限流是防护手段，
			// 不应因为 Redis 抖动把正常用户全部挡在门外。
			global.Logger.Warn("限流判定失败，本次请求放行", zap.String("key", key), zap.Error(err))
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

		if !result.Allowed {
			retryAfter := int(result.RetryAfter.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "访问过于频繁，请稍后重试！",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// newLimiter 按 system.use-redis 选择限流实现。
func newLimiter(rule rateRule) (ratelimit.Limiter, error) {
	if global.Config.System.UseRedis {
		return ratelimit.NewRedisLimiter(global.Redis, rule.limit, rule.window)
	}
	return ratelimit.NewMemoryLimiter(rule.limit, rule.window), nil
}

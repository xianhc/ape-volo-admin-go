package ratelimit

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// NewRedisLimiter 创建共享计数的固定窗口限流器。
//
// 构造时加载计数脚本，失败则返回错误；客户端由缓存生命周期统一关闭。
func NewRedisLimiter(client *redis.Client, limit int64, window time.Duration) (Limiter, error) {
	if client == nil {
		return nil, fmt.Errorf("ratelimit: Redis 客户端未初始化")
	}
	limit, window = normalize(limit, window)
	store, err := redisstore.NewStoreWithOptions(client, limiter.StoreOptions{Prefix: "ratelimit"})
	if err != nil {
		return nil, fmt.Errorf("ratelimit: 初始化 Redis 存储失败: %w", err)
	}
	return newAdapter(store, limit, window, "redis"), nil
}

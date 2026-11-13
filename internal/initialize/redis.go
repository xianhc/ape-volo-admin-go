package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// pingTimeout 是启动时探测 Redis 的超时时间。
// 设短一些，避免 Redis 不可达时进程长时间卡在启动阶段。
const pingTimeout = 5 * time.Second

// Redis 按 system.use-redis 决定缓存与限流后端，并注入 global.Cache。
//
//   - use-redis = false：完全不触碰 Redis，注入进程内存实现。
//   - use-redis = true ：连接 Redis；连接失败视为启动失败（fail fast），
//     不静默切换后端，保证会话、验证码和限流始终使用配置指定的存储。
//
// 返回 error 而不是直接 os.Exit，便于调用方统一处理与测试。
func Redis() error {
	if !global.Config.System.UseRedis {
		global.Cache = cache.NewMemoryStore(time.Minute)
		global.Logger.Info("use-redis=false，缓存与限流使用进程内存实现（仅单实例部署有效）")
		return nil
	}

	redisCfg := global.Config.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		_ = client.Close()
		return fmt.Errorf("use-redis=true 但 Redis(%s) 连接失败: %w", redisCfg.Addr, err)
	}

	global.Redis = client
	global.Cache = cache.NewRedisStore(client)
	global.Logger.Info("redis ping 成功", zap.String("pong", pong), zap.String("addr", redisCfg.Addr))
	return nil
}

// CloseCache 释放缓存后端资源，进程退出前调用。
func CloseCache() {
	if global.Cache == nil {
		return
	}
	if err := global.Cache.Close(); err != nil {
		global.Logger.Warn("关闭缓存后端失败", zap.Error(err))
	}
	global.Cache = nil
	global.Redis = nil
}

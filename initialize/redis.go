package initialize

import (
	"context"

	"go-apevolo/global"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Redis() {
	redisCfg := global.Config.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password, // no password set
		DB:       redisCfg.DB,       // use default DB
	})
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		global.Logger.Error("redis服务连接失败", zap.Error(err))
	} else {
		global.Logger.Info("redis ping 成功", zap.String("pong", pong))
		global.Redis = client
	}
}

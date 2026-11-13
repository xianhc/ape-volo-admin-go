package global

import (
	"github.com/spf13/viper"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/timer"

	"go.uber.org/zap"

	"github.com/xianhc/ape-volo-admin-go/internal/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Db        *gorm.DB
	LogDb     *gorm.DB
	Databases map[string]*gorm.DB
	Redis     *redis.Client
	Config    config.Server
	Viper     *viper.Viper
	Logger    *zap.Logger
	Timer     timer.Timer = timer.NewTimerTask()

	// Cache 是全局缓存抽象，由 initialize.Redis() 根据 system.use-redis 注入：
	//   true  -> Redis 实现（多实例共享）
	//   false -> 进程内存实现（单实例）
	// 业务代码一律通过该接口访问缓存，禁止直接使用 Redis 字段。
	Cache cache.Store
)

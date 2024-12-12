package global

import (
	"github.com/spf13/viper"
	"go-apevolo/utils/timer"

	"go.uber.org/zap"

	"go-apevolo/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Db     *gorm.DB
	Redis  *redis.Client
	Config config.Server
	Viper  *viper.Viper
	Logger *zap.Logger
	Timer  timer.Timer = timer.NewTimerTask()
)

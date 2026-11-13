package internal

import (
	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBBASE interface {
	GetLogMode() string
}

var Gorm = new(_gorm)

type _gorm struct{}

// Config gorm 自定义配置
func (g *_gorm) Config(general config.GeneralDB) *gorm.Config {
	return &gorm.Config{
		Logger: logger.New(NewWriter(general, log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
			SlowThreshold:        200 * time.Millisecond,
			LogLevel:             general.LogLevel(),
			Colorful:             true,
			ParameterizedQueries: true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   general.Prefix,
			SingularTable: general.Singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		DisableAutomaticPing:                     true,
	}
}

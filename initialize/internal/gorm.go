package internal

import (
	"go-apevolo/config"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"go-apevolo/global"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBBASE interface {
	GetLogMode() string
}

var Gorm = new(_gorm)

type _gorm struct{}

// Config gorm 自定义配置
func (g *_gorm) Config(prefix string, singular bool) *gorm.Config {
	var general config.GeneralDB
	switch global.Config.System.DbType {
	case "sqlite":
		general = global.Config.Sqlite.GeneralDB
	case "mysql":
		general = global.Config.Mysql.GeneralDB
	case "pgsql":
		general = global.Config.Pgsql.GeneralDB
	case "oracle":
		general = global.Config.Oracle.GeneralDB
	case "mssql":
		general = global.Config.Mssql.GeneralDB
	default:
		general = global.Config.Sqlite.GeneralDB
	}
	return &gorm.Config{
		Logger: logger.New(NewWriter(general, log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      general.LogLevel(),
			Colorful:      true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   prefix,
			SingularTable: singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
}

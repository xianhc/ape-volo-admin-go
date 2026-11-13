package initialize

import (
	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize/internal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormPgSql 打开预先创建的 PostgreSQL 数据库。
//
// 不读取全局主库配置，也不自动创建数据库。
func GormPgSql(c config.GeneralDB) (*gorm.DB, error) {
	p := config.Pgsql{GeneralDB: c}
	return gorm.Open(postgres.New(postgres.Config{DSN: p.Dsn()}), internal.Gorm.Config(c))
}

package initialize

import (
	"github.com/glebarez/sqlite"
	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize/internal"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

// GormSqlite 用传入配置创建独立 SQLite 文件连接。
//
// 只创建该文件的父目录，连接池和 Ping 由装配层完成。
func GormSqlite(c config.GeneralDB) (*gorm.DB, error) {
	s := config.Sqlite{GeneralDB: c}
	if err := ensureDir(filepath.Dir(s.Dsn())); err != nil {
		return nil, err
	}
	return gorm.Open(sqlite.Open(s.Dsn()), internal.Gorm.Config(c))
}

func ensureDir(dir string) error {
	if dir == "" {
		return nil
	}
	return os.MkdirAll(filepath.Clean(dir), 0o755)
}

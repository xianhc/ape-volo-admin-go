package initialize

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize/internal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormMysql 创建目标库并打开 MySQL 连接。
//
// 管理连接始终关闭，库名使用受控引用而非原文拼接。
func GormMysql(c config.GeneralDB) (*gorm.DB, error) {
	m := config.Mysql{GeneralDB: c}
	admin, err := sql.Open("mysql", m.DsnNotDb())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = admin.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+strings.ReplaceAll(c.Dbname, "`", "``")+"`")
	cancel()
	closeErr := admin.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: m.Dsn(), DefaultStringSize: 255}), internal.Gorm.Config(c))
	if err == nil && c.Engine != "" {
		db = db.Set("gorm:table_options", "ENGINE="+c.Engine).Session(&gorm.Session{})
	}
	return db, err
}

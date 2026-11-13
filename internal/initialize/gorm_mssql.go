package initialize

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize/internal"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// GormMssql 检查或创建目标库并打开 SQL Server 连接。
//
// 管理连接有独立期限，建库标识符按方言转义。
func GormMssql(c config.GeneralDB) (*gorm.DB, error) {
	m := config.Mssql{GeneralDB: c}
	admin, err := sql.Open("sqlserver", m.DsnNotDb())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var count int
	err = admin.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys.databases WHERE name = @p1", c.Dbname).Scan(&count)
	if err == nil && count == 0 {
		_, err = admin.ExecContext(ctx, "CREATE DATABASE ["+strings.ReplaceAll(c.Dbname, "]", "]]")+"]")
	}
	closeErr := admin.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return gorm.Open(sqlserver.New(sqlserver.Config{DSN: m.Dsn(), DefaultStringSize: 255}), internal.Gorm.Config(c))
}

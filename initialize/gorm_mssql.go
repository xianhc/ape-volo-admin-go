package initialize

import (
	"database/sql"
	"fmt"
	"go-apevolo/global"
	"go-apevolo/initialize/internal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// GormMssql 初始化Mssql数据库
func GormMssql() *gorm.DB {
	m := global.Config.Mssql
	if m.Dbname == "" {
		return nil
	}

	// 提取 DSN，不包含数据库名，用于初始连接
	baseDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%s", m.Username, m.Password, m.Host, m.Port)

	// 初始连接（不带数据库名）
	baseDB, err := sql.Open("sqlserver", baseDSN)
	if err != nil {
		global.Logger.Fatal("无法连接到数据库服务器", zap.Error(err))
		return nil
	}
	defer baseDB.Close()

	// 检查数据库是否存在
	var dbExists int
	query := "SELECT COUNT(*) FROM sys.databases WHERE name = @p1"
	err = baseDB.QueryRow(query, m.Dbname).Scan(&dbExists)
	if err != nil {
		global.Logger.Fatal("检查数据库是否存在失败", zap.Error(err))
		return nil
	}

	// 如果数据库不存在，创建它
	if dbExists == 0 {
		createDBSQL := fmt.Sprintf("CREATE DATABASE [%s]", m.Dbname)
		_, err = baseDB.Exec(createDBSQL)
		if err != nil {
			global.Logger.Fatal("创建数据库失败", zap.Error(err))
			return nil
		}
	}

	// 使用完整的 DSN 连接到指定的数据库
	mssqlConfig := sqlserver.Config{
		DSN:               m.Dsn(), // DSN data source name
		DefaultStringSize: 255,     // string 类型字段的默认长度
	}
	db, err := gorm.Open(sqlserver.New(mssqlConfig), internal.Gorm.Config(m.Prefix, m.Singular))
	if err != nil {
		global.Logger.Fatal("连接数据库失败", zap.Error(err))
		return nil
	}

	// 设置 GORM 的其他配置
	db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)

	return db
}

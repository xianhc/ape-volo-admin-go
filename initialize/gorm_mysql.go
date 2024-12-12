package initialize

import (
	_ "github.com/go-sql-driver/mysql"
	"go-apevolo/global"
	"go-apevolo/initialize/internal"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormMysql 初始化Mysql数据库
func GormMysql() *gorm.DB {
	m := global.Config.Mysql
	if m.Dbname == "" {
		return nil
	}

	// 创建 MySQL 配置实例，不指定数据库
	mysqlConfig := mysql.Config{
		DSN:                       m.DsnNotDb(), // 通过不指定数据库名来连接 MySQL 服务
		DefaultStringSize:         255,          // string 类型字段的默认长度
		SkipInitializeWithVersion: false,        // 根据版本自动配置
	}

	// 连接到 MySQL 服务（不连接具体的数据库）
	db, err := gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config(m.Prefix, m.Singular))
	if err != nil {
		global.Logger.Fatal("无法连接到 MySQL 服务:", zap.Error(err))
		return nil
	}

	// 检查数据库是否存在
	if !databaseExists(db, m.Dbname) {
		// 如果数据库不存在，创建数据库
		if err := createDatabase(db, m.Dbname); err != nil {
			global.Logger.Fatal("创建数据库失败:", zap.Error(err))
			return nil
		}
	}

	// 连接到指定的数据库
	mysqlConfig = mysql.Config{
		DSN:                       m.Dsn(), // 使用带有数据库名的DSN连接
		DefaultStringSize:         255,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}

	db, err = gorm.Open(mysql.New(mysqlConfig), internal.Gorm.Config(m.Prefix, m.Singular))
	if err != nil {
		global.Logger.Fatal("无法连接到目标数据库:", zap.Error(err))
		return nil
	}

	// 设置数据库连接选项
	db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)

	return db
}

// 检查数据库是否存在
func databaseExists(db *gorm.DB, dbName string) bool {
	var count int64
	result := db.Raw("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&count)
	if result.Error != nil {
		global.Logger.Fatal("查询数据库存在性失败:", zap.Error(result.Error))
	}
	return count > 0
}

// 创建数据库
func createDatabase(db *gorm.DB, dbName string) error {
	// 使用 Raw 执行 SQL 创建数据库
	result := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	return result.Error
}

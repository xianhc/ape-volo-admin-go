package initialize

import (
	//"github.com/dzwvip/oracle"
	//_ "github.com/godror/godror"
	"go-apevolo/global"
	"go-apevolo/initialize/internal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormOracle 初始化oracle数据库
func GormOracle() *gorm.DB {
	m := global.Config.Oracle
	if m.Dbname == "" {
		return nil
	}
	oracleConfig := mysql.Config{
		DSN:               m.Dsn(), // DSN data source name
		DefaultStringSize: 255,     // string 类型字段的默认长度
	}
	if db, err := gorm.Open(mysql.New(oracleConfig), internal.Gorm.Config(m.Prefix, m.Singular)); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}

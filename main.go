package main

import (
	"database/sql"
	"go-apevolo/core"
	"go-apevolo/global"
	"go-apevolo/initialize"
	"go.uber.org/zap"
	"os"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

func main() {
	global.Viper = core.Viper() // 初始化Vipers
	global.Logger = core.Zap()  // 初始化zap日志库
	zap.ReplaceGlobals(global.Logger)
	global.Db = initialize.Gorm() // gorm连接数据库
	if global.Db != nil {
		// 程序结束前关闭数据库链接
		db, _ := global.Db.DB()
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
	} else {
		global.Logger.Error("应用程序启动失败,未能正确连接数据库")
		os.Exit(0)
	}

	core.Run()
}

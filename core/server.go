package core

import (
	"fmt"
	"go-apevolo/global"
	"go-apevolo/initialize"
	"go-apevolo/utils"
	"net/http"
	"os"
	"time"
)

func Run() {
	// 初始化系统表与数据
	if global.Config.System.IsInitTable {
		err := initialize.InitTable()
		if err != nil {
			os.Exit(0)
		}
		if global.Config.System.IsInitTableData {
			err = initialize.InitTableData()
			if err != nil {
				os.Exit(0)
			}
		}
	}
	// 初始化IP数据
	utils.InitIpData()
	// 初始化redis
	initialize.Redis()
	//初始化定时任务
	initialize.Timer()
	//初始化路由
	router := initialize.Routers()

	port := fmt.Sprintf(":%d", global.Config.System.Port)
	httpServer := &http.Server{
		Addr:           port,
		Handler:        router,
		ReadTimeout:    200 * time.Second,
		WriteTimeout:   200 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf(`应用程序启动成功!{端口号 %s }`, port)
	fmt.Printf(`
欢迎使用《ape-volo-admin》中后台权限管理系统
加群方式:微信号：apevolo<备注'go'>   QQ群：1015661568
项目在线文档:http://doc.apevolo.com/
接口文档地址:http://localhost%s/swagger/api/index.html
前端运行地址:http://localhost:8001
如果项目让您获得了收益，希望您能请作者喝杯咖啡:http://doc.apevolo.com/donate
`, port)
	global.Logger.Error(httpServer.ListenAndServe().Error())
}

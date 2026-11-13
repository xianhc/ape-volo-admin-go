package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

const shutdownTimeout = 30 * time.Second

// Run 按依赖顺序启动服务，退出时先结束 HTTP 和任务，再释放缓存。
func Run(ctx context.Context) (err error) {
	if global.Config.System.IsInitTable {
		if err := initialize.InitTable(); err != nil {
			return err
		}
		if global.Config.System.IsInitTableData {
			if err := initialize.InitTableData(); err != nil {
				return err
			}
		}
	}
	if err := initialize.InitLogTables(ctx); err != nil {
		return fmt.Errorf("初始化日志月表失败: %w", err)
	}
	if err := initialize.StartLogWriters(); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, initialize.StopLogWriters()) }()
	if err := utils.InitIpData(); err != nil {
		return fmt.Errorf("初始化 IP 归属地失败: %w", err)
	}
	if err := initialize.Redis(); err != nil {
		return err
	}
	defer initialize.CloseCache()
	runCtx, stop := context.WithCancel(ctx)
	jobsDone := make(chan error, 1)
	go func() {
		<-runCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		jobsDone <- global.Timer.Shutdown(shutdownCtx)
	}()
	defer func() { stop(); err = errors.Join(err, <-jobsDone) }()
	if err := initialize.Timer(); err != nil {
		return fmt.Errorf("初始化任务失败: %w", err)
	}
	if ctx.Err() != nil {
		return nil
	}
	router, closeRouter, err := initialize.Routers()
	if err != nil {
		return err
	}
	defer closeRouter()
	port := fmt.Sprintf(":%d", global.Config.System.Port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	requestCtx, cancelRequests := context.WithCancel(context.Background())
	defer cancelRequests()
	server := &http.Server{
		Addr: port, Handler: router, ReadTimeout: 200 * time.Second, WriteTimeout: 200 * time.Second,
		MaxHeaderBytes: 1 << 20, BaseContext: func(net.Listener) context.Context { return requestCtx },
	}
	documentAddress := ""
	if global.Config.Swagger.Enabled {
		documentAddress = fmt.Sprintf("接口文档地址：http://localhost%s/swagger/api/index.html\n", port)
	}
	fmt.Printf(`欢迎使用《ape-volo-admin》中后台权限管理系统
当前版本：v%s
加群方式：微信号：apevolo<备注'加群'>   QQ群：839263566
官方网站：https://www.apevolo.com
%s前端运行地址：http://localhost:9001
☕ 请作者喝杯咖啡，支持项目持续维护：https://www.apevolo.com/donate
+---------------------------------- 版权声明 ----------------------------------+
|              ** 版权所有 © 2022–%d xianhc。保留所有权利。 **               |
|    ** 开源许可：Apache-2.0，详见 LICENSE；联系邮箱：apevolo@gmail.com **     |
+------------------------------------------------------------------------------+
`, global.Config.Swagger.Version, documentAddress, time.Now().Year())
	return serveHTTP(runCtx, server, listener, cancelRequests)
}

func serveHTTP(ctx context.Context, server *http.Server, listener net.Listener, cancelRequests context.CancelFunc) error {
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		cancelRequests()
		if shutdownErr != nil {
			_ = server.Close()
		}
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return errors.Join(err, shutdownErr)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			cancelRequests()
			_ = server.Close()
		}
		serveErr := <-done
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		return errors.Join(err, serveErr)
	}
}

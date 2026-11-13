package initialize

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
	"github.com/xianhc/ape-volo-admin-go/internal/router"
)

// Routers 初始化总路由
func Routers() (*gin.Engine, func(), error) {
	// 路由注册输出也换成项目前缀，避免启动日志里出现 [GIN-debug]。
	// 该函数必须在注册任何路由之前设置。
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		fmt.Printf("[ROUTE] %-6s %-45s (%d handlers)\n", httpMethod, absolutePath, nuHandlers)
	}

	// 不用 gin.Default()：它内置的 Logger 会打上 [GIN] 前缀。
	// 这里用 gin.New() + 自定义 Formatter，统一成项目自己的访问日志格式。
	Router := gin.New()
	Router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// 与 Gin 默认 Formatter 保持一致：状态码、耗时、方法都按终端能力上色，
			// 耗时还按量级截断，避免长尾请求把整行撑开。
			var statusColor, methodColor, resetColor, latencyColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
				latencyColor = param.LatencyColor()
			}

			switch {
			case param.Latency > time.Minute:
				param.Latency = param.Latency.Truncate(time.Second * 10)
			case param.Latency > time.Second:
				param.Latency = param.Latency.Truncate(time.Millisecond * 10)
			case param.Latency > time.Millisecond:
				param.Latency = param.Latency.Truncate(time.Microsecond * 10)
			}

			return fmt.Sprintf("[APE-VOLO] %v |%s %3d %s|%s %8v %s| %15s |%s %-7s %s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, param.StatusCode, resetColor,
				latencyColor, param.Latency, resetColor,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
				param.ErrorMessage,
			)
		},
	}))
	Router.Use(gin.Recovery())

	// 受信代理配置：决定 c.ClientIP() 是否采信 X-Forwarded-For。
	// 不配置时 gin 默认信任所有代理，等于允许客户端伪造 IP，
	// 会绕过限流并污染审计日志与登录地记录。
	// 未配置 trusted-proxies 时显式关闭代理信任，只取 TCP 对端地址。
	trustedProxies := global.Config.System.TrustedProxies
	if len(trustedProxies) == 0 {
		if err := Router.SetTrustedProxies(nil); err != nil {
			global.Logger.Warn("关闭受信代理失败: " + err.Error())
		}
	} else if err := Router.SetTrustedProxies(trustedProxies); err != nil {
		global.Logger.Warn("设置受信代理失败: " + err.Error())
	}

	authorizationRouter := router.RouterGroupApp.Authorization

	Router.StaticFS(global.Config.Local.Path, http.Dir(global.Config.Local.StorePath))
	Router.Use(middleware.RequestID())
	rateLimit, closeRateLimit, err := middleware.IpLimit()
	if err != nil {
		return nil, nil, err
	}
	Router.Use(rateLimit)
	// Router.Use(middleware.LoadTls())  // 如果需要使用https 请打开此中间件 然后前往 core/server.go 将启动模式 更变为 Router.RunTLS("端口","你的cre/pem文件","你的key文件")
	// 跨域，按照配置的规则放行跨域请求
	Router.Use(middleware.CorsByRules())
	registerSwagger(Router)

	// 健康检查：`/health/index` 匿名访问并返回固定成功标记。
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, response.ActionResult{IsSuccess: true, Status: http.StatusOK})
	}
	Router.GET("/health/index", healthHandler)

	PublicGroup := Router.Group(global.Config.System.RouterPrefix)
	{
		// 验证码、登录、刷新令牌：匿名可访问，但与其他接口同属 `/api/auth/*`
		authorizationRouter.InitAuthorizationRouterAllowAnonymous(PublicGroup)
	}
	PrivateGroup := Router.Group(global.Config.System.RouterPrefix)
	PrivateGroup.Use(middleware.JWTAuth()).Use(middleware.PermissionHandler()).Use(middleware.Transaction())
	privateRoutes := routeauth.NewGroup(PrivateGroup)
	{
		router.RouterGroupApp.InitPrivate(privateRoutes)
	}
	// 启动自检：routeauth 声明的路由必须真实存在。
	validatePermissionRoutes(Router)
	return Router, closeRateLimit, nil
}

// validatePermissionRoutes 校验 routeauth 声明的私有路由均已注册到 Gin。
//
// 反向遗漏由 PermissionHandler 在运行期 fail closed；routeauth.Group 正常使用时两者由同一次调用产生。
func validatePermissionRoutes(r *gin.Engine) {
	prefix := "/" + strings.Trim(global.Config.System.RouterPrefix, "/")

	registered := make(map[string]struct{})
	for _, route := range r.Routes() {
		path := route.Path
		if prefix != "/" {
			path = strings.TrimPrefix(path, prefix)
		}
		registered[route.Method+" "+path] = struct{}{}
	}

	for _, rule := range routeauth.Rules() {
		if _, ok := registered[rule.Method+" "+rule.Path]; !ok {
			global.Logger.Warn("授权策略对应的路由不存在，请检查注册过程: " + rule.Method + " " + rule.Path)
		}
	}
}

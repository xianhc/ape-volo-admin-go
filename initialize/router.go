package initialize

import (
	"go-apevolo/docs"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-apevolo/global"
	"go-apevolo/middleware"
	"go-apevolo/router"
)

// Routers 初始化总路由
func Routers() *gin.Engine {
	Router := gin.Default()
	systemRouter := router.RouterGroupApp.System
	permissionRouter := router.RouterGroupApp.Permission
	authorizationRouter := router.RouterGroupApp.Authorization
	monitorRouter := router.RouterGroupApp.Monitor
	messageRouter := router.RouterGroupApp.Message
	queuedRouter := router.RouterGroupApp.Queued

	Router.StaticFS(global.Config.Local.StorePath, http.Dir(global.Config.Local.StorePath))
	Router.Use(middleware.IpLimit())
	// Router.Use(middleware.LoadTls())  // 如果需要使用https 请打开此中间件 然后前往 core/server.go 将启动模式 更变为 Router.RunTLS("端口","你的cre/pem文件","你的key文件")
	// 跨域，按照配置的规则放行跨域请求
	Router.Use(middleware.CorsByRules())
	docs.SwaggerInfo.BasePath = ""
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	PublicGroupNotRouterPrefix := Router.Group("")
	{
		authorizationRouter.InitAuthorizationRouterAllowAnonymous(PublicGroupNotRouterPrefix) //验证码 登录 刷新token
	}

	PublicGroup := Router.Group(global.Config.System.RouterPrefix)
	{
		// 健康监测
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})
	}
	PrivateGroup := Router.Group(global.Config.System.RouterPrefix)
	PrivateGroupNotRouterPrefix := Router.Group("")
	PrivateGroup.Use(middleware.JWTAuth()).Use(middleware.PermissionHandler())
	{

		authorizationRouter.InitAuthorizationRouter(PrivateGroupNotRouterPrefix) //个人信息
		permissionRouter.InitUserRouter(PrivateGroup)                            //用户
		permissionRouter.InitMenuRouter(PrivateGroup)                            //菜单
		permissionRouter.InitDeptRouter(PrivateGroup)                            //部门
		permissionRouter.InitJobRouter(PrivateGroup)                             //岗位
		permissionRouter.InitRoleRouter(PrivateGroup)                            //角色
		permissionRouter.InitApisRouter(PrivateGroup)                            //Apis
		permissionRouter.InitRolePermissionRouter(PrivateGroup)                  //菜单 路由

		systemRouter.InitDictRouter(PrivateGroup)       //字典
		systemRouter.InitDictDetailRouter(PrivateGroup) //字典详情
		systemRouter.InitSettingRouter(PrivateGroup)    // 全局设置
		systemRouter.InitAppSecretRouter(PrivateGroup)  // 应用密钥
		systemRouter.InitFileRecordRouter(PrivateGroup) //文件记录
		systemRouter.InitTaskRouter(PrivateGroup)       //任务调度

		monitorRouter.InitOnlineUserRouter(PrivateGroup)      //在线用户
		monitorRouter.InitAuditLogRouter(PrivateGroup)        //审计日志
		monitorRouter.InitExceptionLogRouter(PrivateGroup)    //异常日志
		monitorRouter.InitServerResourcesRouter(PrivateGroup) //服务器资源信息

		messageRouter.InitEmailAccountRouterWithoutRecord(PrivateGroup)    //邮箱账户
		messageRouter.InitMessageTemplateRouterWithoutRecord(PrivateGroup) //邮件模板

		queuedRouter.InitEmailQueuedRouterWithoutRecord(PrivateGroup) //邮件队列

	}
	return Router
}

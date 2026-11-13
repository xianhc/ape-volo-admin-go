package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/authorization/auth"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type AuthorizationRouter struct{}

func (s *AuthorizationRouter) InitAuthorizationRouterAllowAnonymous(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("auth", middleware.LogModule("authorization", "auth"))
	authHandler := auth.Handler{}
	{
		baseRouter.GET("captcha", authHandler.Captcha)
		baseRouter.POST("login", middleware.LogAction("用户登录"), middleware.OperationRecord(), authHandler.Login)
		baseRouter.POST("refresh-token", authHandler.RefreshToken)
		baseRouter.DELETE("logout", middleware.LogAction("用户退出"), middleware.OperationRecord(), authHandler.Logout)
	}
	return baseRouter
}

func (s *AuthorizationRouter) InitAuthorizationRouter(Router *routeauth.Group) {
	baseRouter := Router.Group("auth", middleware.LogModule("authorization", "auth"), middleware.OperationRecord())
	authHandler := auth.Handler{}
	{
		baseRouter.GET("info", routeauth.LoginOnly(), middleware.LogAction("获取当前用户信息"), authHandler.GetInfo)
		baseRouter.POST("reset-email-code", routeauth.LoginOnly(), middleware.LogAction("发送邮箱验证码"), middleware.OperationRecord(), authHandler.ResetEmail)
	}
}

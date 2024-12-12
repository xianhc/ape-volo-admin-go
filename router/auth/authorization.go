package auth

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/api"
	"go-apevolo/middleware"
)

type AuthorizationRouter struct{}

func (s *AuthorizationRouter) InitAuthorizationRouterAllowAnonymous(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("auth")
	authApi := api.ApiGroupApp.AuthorizationApiGroup.AuthorizationApi
	{
		baseRouter.GET("captcha", authApi.Captcha)                            //获取验证码
		baseRouter.POST("login", middleware.OperationRecord(), authApi.Login) //登录
		baseRouter.POST("refreshToken", authApi.RefreshToken)                 //刷新token
	}
	return baseRouter
}

func (s *AuthorizationRouter) InitAuthorizationRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("auth").Use(middleware.OperationRecord())
	authApi := api.ApiGroupApp.AuthorizationApiGroup.AuthorizationApi
	{
		baseRouter.GET("info", authApi.GetInfo)                                                //个人信息
		baseRouter.DELETE("logout", middleware.OperationRecord(), authApi.Logout)              //登出
		baseRouter.POST("/code/reset/email", middleware.OperationRecord(), authApi.ResetEmail) //重置邮箱验证码
	}
	return baseRouter
}

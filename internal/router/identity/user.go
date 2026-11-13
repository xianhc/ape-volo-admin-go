package identity

import (
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/permission"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/rolecode"
	"github.com/xianhc/ape-volo-admin-go/internal/middleware"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
)

type UserRouter struct{}

func (u *UserRouter) InitUserRouter(Router *routeauth.Group) {
	userRouter := Router.Group("user", middleware.LogModule("identity", "user"), middleware.OperationRecord())
	userHandler := user.Handler{}
	{
		userRouter.POST("create", routeauth.Require(permission.SysUserAdd, rolecode.Admin), middleware.LogAction("创建用户"), userHandler.Create)
		userRouter.PUT("update", routeauth.Require(permission.SysUserEdit), middleware.LogAction("修改用户"), userHandler.Update)
		userRouter.DELETE("delete", routeauth.Require(permission.SysUserDelete), middleware.LogAction("删除用户"), userHandler.Delete)
		userRouter.GET("get-list", routeauth.Require(permission.SysUserList), middleware.LogAction("查询用户列表"), userHandler.Query)
		userRouter.POST("download", routeauth.Require(permission.SysUserDownload), middleware.LogAction("导出用户"), userHandler.Download)

		userRouter.PUT("update-center", routeauth.LoginOnly(), middleware.LogAction("修改个人资料"), userHandler.UpdateCenter)
		userRouter.PUT("update-password", routeauth.LoginOnly(), middleware.LogAction("修改密码"), userHandler.UpdatePassword)
		userRouter.PUT("update-email", routeauth.LoginOnly(), middleware.LogAction("修改邮箱"), userHandler.UpdateEmail)
		userRouter.PUT("update-position", routeauth.Require(permission.SysUserEdit), middleware.LogAction("修改用户岗位"), userHandler.UpdatePositions)
		userRouter.PUT("update-role", routeauth.Require(permission.SysUserEdit, rolecode.Admin), middleware.LogAction("修改用户角色"), userHandler.UpdateRoles)
		userRouter.PUT("update-preferences-config", routeauth.LoginOnly(), middleware.LogAction("修改界面偏好"), userHandler.UpdatePreferences)
		userRouter.POST("update-avatar", routeauth.LoginOnly(), middleware.LogAction("修改头像"), userHandler.UpdateAvatar)
	}
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	sessionmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/authorization/session"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/role"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/auth"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/routeauth"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/querycache"
)

var rolePermissionService = role.Service{}

// PermissionHandler 权限拦截处理器。
//
// 判定链路：token → 服务端会话 → 路由声明的角色/权限码 → 用户授权集合。
// 权限策略由 routeauth.Group 在注册路由时同步登记，不维护第二份路由映射表。
func PermissionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		message, code := permissionHandler(c)
		if code == http.StatusOK {
			c.Next()
			return
		}

		switch code {
		case http.StatusUnauthorized:
			response.Unauthorized(message, nil, c)
		case http.StatusForbidden:
			response.Forbidden(message, nil, c)
		default:
			response.Error(message, nil, c)
		}
		c.Abort()
	}
}

func permissionHandler(c *gin.Context) (message string, httpCode int) {
	ctx := c.Request.Context()
	userId := utils.GetId(c)

	// 1. 校验会话
	loginUserInfo, err := resolveSession(c)
	if err != nil {
		return err.Error(), http.StatusBadRequest
	}
	if loginUserInfo == nil {
		return "抱歉，您无权访问该接口", http.StatusUnauthorized
	}
	roleCodes, err := loadRoleCodes(ctx, userId)
	if err != nil {
		return err.Error(), http.StatusBadRequest
	}
	ctx = utils.WithRoleCodes(ctx, roleCodes)
	c.Request = c.Request.WithContext(ctx)

	// 2. 查该路由声明的权限码
	policy, registered := routeauth.Lookup(c.Request.Method, routePath(c))
	if !registered {
		// 未登记的路由一律拒绝（fail closed），并留下明显日志便于补齐契约表
		global.Logger.Error("路由未登记权限码，已拒绝访问: " + c.Request.Method + " " + c.FullPath())
		return "接口未配置权限，已拒绝访问", http.StatusForbidden
	}
	for _, requiredRole := range policy.Roles() {
		if !utils.HasRole(ctx, requiredRole) {
			return "抱歉，您访问权限等级不够", http.StatusForbidden
		}
	}
	if policy.Code() == "" {
		// 仅需登录，无需具体权限
		return "", http.StatusOK
	}

	// 3. 权限码校验
	codes, err := loadPermissionCodes(ctx, userId)
	if err != nil {
		return err.Error(), http.StatusBadRequest
	}
	for _, code := range codes {
		if code == policy.Code() {
			return "", http.StatusOK
		}
	}
	return "抱歉，您访问权限等级不够", http.StatusForbidden
}

// routePath 返回剥离 router-prefix 后的路由模板，供契约表匹配。
//
// 例如 router-prefix 为 api 时，`/api/user/query` → `/user/query`；
// 而注册在根路径下的 `/auth/info` 保持不变。
func routePath(c *gin.Context) string {
	path := c.FullPath()
	prefix := "/" + strings.Trim(global.Config.System.RouterPrefix, "/")
	if prefix != "/" {
		path = strings.TrimPrefix(path, prefix)
	}
	return path
}

// resolveSession 校验令牌对应的会话是否仍然有效。
//
// 会话是认证的唯一依据：缓存里没有就视为未登录，不从数据库重建。
// 重建会让「踢下线」「改密码后强制重登」失去意义——只要令牌没过期就还能用。
func resolveSession(c *gin.Context) (*auth.LoginUserInfo, error) {
	value, exists := c.Get("claims")
	if !exists {
		return nil, nil
	}
	claims, ok := value.(*utils.ApeClaims)
	if !ok || claims == nil || claims.Sid == "" {
		return nil, nil
	}

	session, err := (&sessionmodule.Service{}).FindActive(c.Request.Context(), claims.Sid)
	if err != nil || session == nil {
		return nil, err
	}
	// 会话必须属于令牌声明的用户
	if session.UserId != utils.UserID(claims) {
		return nil, nil
	}
	// 刷新后旧访问令牌立即失效
	if session.AccessTokenId != claims.ID {
		return nil, nil
	}
	return session.LoginUserInfo()
}

// loadPermissionCodes 读取用户的权限码集合，结果按用户缓存 1 小时。
func loadPermissionCodes(ctx context.Context, userId int64) ([]string, error) {
	codes := &[]string{}
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserAuthCodes, userId), codes,
		ext.GetTimeDuration(1, ext.Hour),
		func() error {
			return rolePermissionService.GetPermissionIdentifier(ctx, userId, []int64{}, codes)
		})
	if err != nil {
		return nil, err
	}
	return *codes, nil
}

func loadRoleCodes(ctx context.Context, userId int64) ([]string, error) {
	codes := &[]string{}
	err := querycache.GetOrLoad(ctx, querycache.Key(cachePrefix.UserRoleCodes, userId), codes,
		ext.GetTimeDuration(1, ext.Hour),
		func() error {
			var err error
			*codes, err = rolePermissionService.RoleCodes(ctx, userId)
			return err
		})
	if err != nil {
		return nil, err
	}
	return *codes, nil
}

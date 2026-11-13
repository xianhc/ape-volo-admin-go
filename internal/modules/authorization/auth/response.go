package auth

import (
	"github.com/xianhc/ape-volo-admin-go/internal/modules/identity/user"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// AuthContext 表示当前用户认证上下文。
//
// 用户、角色码和按钮权限码平铺在响应顶层。
type AuthContext struct {
	User      user.UserResponse `json:"user"`
	RoleCodes []string          `json:"roleCodes"`
	AuthCodes []string          `json:"authCodes"`
}

// Login 表示登录成功结果。
//
// 响应包含脱敏用户、角色码、权限码和访问/刷新令牌对。
type Login struct {
	User      user.UserResponse `json:"user"`
	RoleCodes []string          `json:"roleCodes"`
	AuthCodes []string          `json:"authCodes"`
	Token     utils.JwtToken    `json:"token"`
}

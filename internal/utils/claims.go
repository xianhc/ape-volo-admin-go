package utils

import "github.com/golang-jwt/jwt/v5"

// 令牌用途。访问令牌与刷新令牌互不通用，防止用刷新凭证访问接口。
const (
	// TokenUseAccess 访问令牌
	TokenUseAccess = "access"
	// TokenUseRefresh 刷新令牌
	TokenUseRefresh = "refresh"
)

// ApeClaims 令牌声明。
//
// 用户身份放在标准字段 sub 里，会话标识放在 sid：
// 会话才是认证依据，令牌只是会话的一个凭证。
//
// 放在 utils 而不是 payloads/request：它是签发与校验令牌的基础类型，
// 让工具包反过来依赖传输对象会把依赖方向拧成环。
type ApeClaims struct {
	// Sid 会话标识
	Sid string `json:"sid"`
	// TokenUse 令牌用途，access 或 refresh
	TokenUse string `json:"token_use"`
	// Name 账号名，仅用于日志与审计，不参与鉴权
	Name string `json:"name"`
	// DeptId 部门主键
	DeptId int64 `json:"dept_id,string"`
	jwt.RegisteredClaims
}

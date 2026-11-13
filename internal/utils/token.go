package utils

// JwtToken 登录成功后返回的令牌对。
//
// 由 CreateToken 产出，因此定义在 utils；响应结构直接复用它，
// 避免工具包为了一个返回类型去依赖 payloads。
type JwtToken struct {
	// SessionId 本次登录的会话标识，不可用于认证
	SessionId string `json:"sessionId"`
	// AccessToken 访问令牌
	AccessToken string `json:"accessToken"`
	// Expires 访问令牌过期时刻，Unix 毫秒时间戳
	Expires int64 `json:"expires"`
	// TokenType 令牌类型
	TokenType string `json:"tokenType"`
	// RefreshToken 刷新令牌，每次刷新后必须保存新值，不能用于访问接口
	RefreshToken string `json:"refreshToken"`
	// RefreshTokenExpires 会话绝对截止时刻，Unix 毫秒时间戳，刷新不会延长
	RefreshTokenExpires int64 `json:"refreshTokenExpires"`
}

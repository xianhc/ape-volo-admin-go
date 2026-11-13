package config

// JwtAuthOptions 令牌签发配置。
type JwtAuthOptions struct {
	Audience string `mapstructure:"audience" json:"audience" yaml:"audience"` // 听众
	Issuer   string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`       // 签发者
	// SecurityKey HS256 签名密钥
	SecurityKey string `mapstructure:"security_key" json:"security_key" yaml:"security_key"`
	// AccessTokenLifetimeMs 访问令牌有效期（毫秒）。
	// 实际签发时取「当前时间 + 本值」与会话绝对截止时间中的较小者。
	AccessTokenLifetimeMs int64 `mapstructure:"access_token_lifetime_ms" json:"access_token_lifetime_ms" yaml:"access_token_lifetime_ms"`
	// RefreshTokenLifetimeMs 刷新令牌有效期（毫秒），同时决定会话的绝对截止时间：
	// 自登录起算，期间刷新只轮换不延长，到期必须重新登录。
	RefreshTokenLifetimeMs int64  `mapstructure:"refresh_token_lifetime_ms" json:"refresh_token_lifetime_ms" yaml:"refresh_token_lifetime_ms"`
	LoginPath              string `mapstructure:"login_path" json:"login_path" yaml:"login_path"` // 登录路径
}

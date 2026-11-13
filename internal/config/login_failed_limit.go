package config

type LoginFailedLimit struct {
	Enabled     bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`                // 是否启用
	MaxAttempts int  `mapstructure:"max_attempts" json:"max_attempts" yaml:"max_attempts"` // 最大尝试次数
	Lockout     int  `mapstructure:"lockout" json:"lockout" yaml:"lockout"`                // 锁定时间 s(秒)
}

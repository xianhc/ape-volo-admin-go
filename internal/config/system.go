package config

type System struct {
	IsInitTable     bool   `mapstructure:"is-init-table" json:"is-init-table" yaml:"is-init-table"`                // 初始化表
	IsInitTableData bool   `mapstructure:"is-init-table-data" json:"is-init-table-data" yaml:"is-init-table-data"` // 初始化表数据
	MainDatabase    string `mapstructure:"main-database" json:"main-database" yaml:"main-database"`                // 主库连接 ID
	OssType         string `mapstructure:"oss-type" json:"oss-type" yaml:"oss-type"`                               // Oss类型
	FileLimitSize   int64  `mapstructure:"file-limit-size" json:"file-limit-size" yaml:"file-limit-size"`          // 上传文件最大限制 单位(mb)
	RouterPrefix    string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Port            int    `mapstructure:"port" json:"port" yaml:"port"`                               // 端口值
	UseRedis        bool   `mapstructure:"use-redis" json:"use-redis" yaml:"use-redis"`                // 使用redis
	IsQuickDebug    bool   `mapstructure:"is-quick-debug" json:"is-quick-debug" yaml:"is-quick-debug"` // 是否开发环境
	// UserDefaultPassword 新建用户的初始密码。
	// 生产环境应通过 APEVOLO_SYSTEM_USER_DEFAULT_PASSWORD 环境变量覆盖。
	UserDefaultPassword string `mapstructure:"user-default-password" json:"user-default-password" yaml:"user-default-password"`
	// TrustedProxies 受信反向代理网段。
	// 只有来自这些地址的请求，其 X-Forwarded-For / X-Real-IP 才会被采信；
	// 为空表示不信任任何代理，直接取 TCP 连接的对端地址（默认，最安全）。
	TrustedProxies []string `mapstructure:"trusted-proxies" json:"trusted-proxies" yaml:"trusted-proxies"`
}

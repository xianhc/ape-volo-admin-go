package config

type System struct {
	IsInitTable     bool   `mapstructure:"is-init-table" json:"is-init-table" yaml:"is-init-table"`                // 初始化表
	IsInitTableData bool   `mapstructure:"is-init-table-data" json:"is-init-table-data" yaml:"is-init-table-data"` // 初始化表数据
	DbType          string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`                                  // 数据库类型:(默认sqlite)
	OssType         string `mapstructure:"oss-type" json:"oss-type" yaml:"oss-type"`                               // Oss类型
	FileLimitSize   int64  `mapstructure:"file-limit-size" json:"file-limit-size" yaml:"file-limit-size"`          // 上传文件最大限制 单位(mb)
	RouterPrefix    string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Port            int    `mapstructure:"port" json:"port" yaml:"port"`                               // 端口值
	UseRedis        bool   `mapstructure:"use-redis" json:"use-redis" yaml:"use-redis"`                // 使用redis
	IsQuickDebug    bool   `mapstructure:"is-quick-debug" json:"is-quick-debug" yaml:"is-quick-debug"` // 是否开发环境
}

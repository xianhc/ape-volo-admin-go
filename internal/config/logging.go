package config

// LogStorage 配置日志月份和统计使用的时区。
//
// 时间解释不依赖宿主机的本地时区。
type LogStorage struct {
	TimeZone string `mapstructure:"time-zone" json:"time-zone" yaml:"time-zone"`
}

// LogQueue 配置每类队列的容量、批次和消费期限。
//
// 三类队列容量独立，关闭时共享总排空期限。
type LogQueue struct {
	Capacity               int `mapstructure:"capacity" json:"capacity" yaml:"capacity"`
	BatchSize              int `mapstructure:"batch-size" json:"batch-size" yaml:"batch-size"`
	WriteTimeoutSeconds    int `mapstructure:"write-timeout-seconds" json:"write-timeout-seconds" yaml:"write-timeout-seconds"`
	ShutdownTimeoutSeconds int `mapstructure:"shutdown-timeout-seconds" json:"shutdown-timeout-seconds" yaml:"shutdown-timeout-seconds"`
}

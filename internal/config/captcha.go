package config

type Captcha struct {
	CodeLength        int `mapstructure:"code-length" json:"code-length" yaml:"code-length"`                         // 验证码长度
	ImageWidth        int `mapstructure:"image-width" json:"image-width" yaml:"image-width"`                         // 图像宽度
	ImageHeight       int `mapstructure:"image-height" json:"image-height" yaml:"image-height"`                      // 图像高度
	FontSize          int `mapstructure:"font-size" json:"font-size" yaml:"font-size"`                               // 字体大小
	FailureThreshold  int `mapstructure:"failure-threshold" json:"failure-threshold" yaml:"failure-threshold"`       // 登录失败次数阈值，超过后才要求验证码
	CounterTtlSeconds int `mapstructure:"counter-ttl-seconds" json:"counter-ttl-seconds" yaml:"counter-ttl-seconds"` // 失败计数器过期时间 s(秒)
}

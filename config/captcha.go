package config

type Captcha struct {
	KeyLength int `mapstructure:"key-length" json:"key-length" yaml:"key-length"` // 验证码长度
	ImgWidth  int `mapstructure:"img-width" json:"img-width" yaml:"img-width"`    // 验证码宽度
	ImgHeight int `mapstructure:"img-height" json:"img-height" yaml:"img-height"` // 验证码高度
	Threshold int `mapstructure:"threshold" json:"threshold" yaml:"threshold"`    // 失败次数阈值
	TimeOut   int `mapstructure:"time-out" json:"time-out" yaml:"time-out"`       // 超时时间 s(秒)
}

package config

type Server struct {
	JwtAuthOptions   JwtAuthOptions   `mapstructure:"jwt-auth-options" json:"jwt-auth-options" yaml:"jwt-auth-options"`
	Zap              Zap              `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis            Redis            `mapstructure:"redis" json:"redis" yaml:"redis"`
	System           System           `mapstructure:"system" json:"system" yaml:"system"`
	Swagger          Swagger          `mapstructure:"swagger" json:"swagger" yaml:"swagger"`
	Captcha          Captcha          `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	LoginFailedLimit LoginFailedLimit `mapstructure:"login-failed-limit" json:"login-failed-limit" yaml:"login-failed-limit"`
	Aop              Aop              `mapstructure:"aop" json:"aop" yaml:"aop"`

	DataConnection    DataConnection `mapstructure:"data-connection" json:"data-connection" yaml:"data-connection"`
	LogDataConnection Connection     `mapstructure:"log-data-connection" json:"log-data-connection" yaml:"log-data-connection"`
	LogStorage        LogStorage     `mapstructure:"log-storage" json:"log-storage" yaml:"log-storage"`
	LogQueue          LogQueue       `mapstructure:"log-queue" json:"log-queue" yaml:"log-queue"`

	Local Local `mapstructure:"local" json:"local" yaml:"local"`

	Excel Excel `mapstructure:"excel" json:"excel" yaml:"excel"`

	// 跨域配置
	Cors CORS `mapstructure:"cors" json:"cors" yaml:"cors"`

	// Rsa
	Rsa Rsa `mapstructure:"rsa" json:"rsa" yaml:"rsa"`
}

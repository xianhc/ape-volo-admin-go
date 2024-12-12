package config

type Server struct {
	JwtAuthOptions   JwtAuthOptions   `mapstructure:"jwt-auth-options" json:"jwt-auth-options" yaml:"jwt-auth-options"`
	Zap              Zap              `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis            Redis            `mapstructure:"redis" json:"redis" yaml:"redis"`
	System           System           `mapstructure:"system" json:"system" yaml:"system"`
	Captcha          Captcha          `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	LoginFailedLimit LoginFailedLimit `mapstructure:"login-failed-limit" json:"login-failed-limit" yaml:"login-failed-limit"`

	// gorm
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mssql  Mssql  `mapstructure:"mssql" json:"mssql" yaml:"mssql"`
	Pgsql  Pgsql  `mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
	Oracle Oracle `mapstructure:"oracle" json:"oracle" yaml:"oracle"`
	Sqlite Sqlite `mapstructure:"sqlite" json:"sqlite" yaml:"sqlite"`
	//DBList []SpecializedDB `mapstructure:"db-list" json:"db-list" yaml:"db-list"`

	Local Local `mapstructure:"local" json:"local" yaml:"local"`

	Excel Excel `mapstructure:"excel" json:"excel" yaml:"excel"`

	// 跨域配置
	Cors CORS `mapstructure:"cors" json:"cors" yaml:"cors"`

	// Rsa
	Rsa Rsa `mapstructure:"rsa" json:"rsa" yaml:"rsa"`
}

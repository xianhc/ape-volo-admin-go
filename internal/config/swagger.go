package config

// Swagger 配置接口文档的开关、名称和展示信息。
//
// 配置在启动时生效，关闭后不注册文档页面和 JSON 路由。
type Swagger struct {
	Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Name         string `mapstructure:"name" json:"name" yaml:"name"`
	Version      string `mapstructure:"version" json:"version" yaml:"version"`
	Title        string `mapstructure:"title" json:"title" yaml:"title"`
	ContactName  string `mapstructure:"contact-name" json:"contact-name" yaml:"contact-name"`
	ContactEmail string `mapstructure:"contact-email" json:"contact-email" yaml:"contact-email"`
	ContactURL   string `mapstructure:"contact-url" json:"contact-url" yaml:"contact-url"`
}

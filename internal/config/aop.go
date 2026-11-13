package config

// Aop 面向切面配置。
//
// 事务与缓存切面共用同一配置节点。
type Aop struct {
	// Transactions 事务切面开关。true 时写请求由中间件统一开启事务；
	// false 时由 service 管理事务。
	Transactions bool `mapstructure:"transactions" json:"transactions" yaml:"transactions"`
	// Cache 查询缓存开关。false 时 querycache.GetOrLoad 直接透传，不读写缓存
	Cache bool `mapstructure:"cache" json:"cache" yaml:"cache"`
}

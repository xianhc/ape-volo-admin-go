package config

type Rsa struct {
	PrivateKey string `mapstructure:"privateKey" json:"privateKey" yaml:"privateKey"` // 私钥
	PublicKey  string `mapstructure:"publicKey" json:"publicKey" yaml:"publicKey"`    // 公钥
}

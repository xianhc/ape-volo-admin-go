package config

// Rsa 密码传输加密配置。
//
// 密码在传输前由前端用 RSA 公钥加密，服务端用私钥解密。
// 私钥必须是 4096 位 PKCS#8 PEM，配合 OAEP-SHA256 填充：
// PKCS#1 v1.5 存在 Bleichenbacher 类填充预言攻击面，OAEP 是当前的推荐用法。
type Rsa struct {
	// PrivateKeyPath PKCS#8 PEM 私钥文件路径，相对路径以进程工作目录为基准。
	PrivateKeyPath string `mapstructure:"private-key-path" json:"privateKeyPath" yaml:"private-key-path"`

	// PrivateKey 直接内联 PEM 内容，优先级高于 PrivateKeyPath。
	// 容器部署时用环境变量注入，避免把密钥文件打进镜像。
	PrivateKey string `mapstructure:"private-key" json:"privateKey" yaml:"private-key"`
}

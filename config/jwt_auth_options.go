package config

type JwtAuthOptions struct {
	Audience            string `mapstructure:"audience" json:"audience" yaml:"audience"`                                        //听众
	Issuer              string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`                                              //签发者
	SecurityKey         string `mapstructure:"security_key" json:"security_key" yaml:"security_key"`                            //签名
	Expires             int32  `mapstructure:"expires" json:"expires" yaml:"expires"`                                           //过期时间
	RefreshTokenExpires int32  `mapstructure:"refresh_token_expires" json:"refresh_token_expires" yaml:"refresh_token_expires"` //刷新时间
	LoginPath           string `mapstructure:"login_path" json:"login_path" yaml:"login_path"`                                  //登录路径
}

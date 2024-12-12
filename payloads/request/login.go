package request

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginAuthUser struct {
	Username  string `json:"username" validate:"required"`  // 用户名
	Password  string `json:"password" validate:"required"`  // 密码
	Captcha   string `json:"captcha"`                       // 验证码
	CaptchaId string `json:"captchaId" validate:"required"` // 验证码ID
}

type ApeClaims struct {
	Claims
	jwt.RegisteredClaims
}

type Claims struct {
	Jti  int64
	Name string
	Iat  int64
	Ip   string
}

type RefreshToken struct {
	Token string `json:"token" validate:"required"` // 用户名
}

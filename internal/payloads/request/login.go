package request

type LoginAuthUser struct {
	UserName  string `json:"userName" validate:"required"`  // 用户名
	Password  string `json:"password" validate:"required"`  // 密码
	Captcha   string `json:"captcha"`                       // 验证码
	CaptchaId string `json:"captchaId" validate:"required"` // 验证码ID
}

type RefreshToken struct {
	Token string `json:"token" validate:"required"` // 用户名
}

type EmailCode struct {
	Email string `json:"email" validate:"required,email"`
}

package response

type Captcha struct {
	Img         string `json:"img"`
	CaptchaId   string `json:"captchaId"`
	ShowCaptcha bool   `json:"showCaptcha"`
}

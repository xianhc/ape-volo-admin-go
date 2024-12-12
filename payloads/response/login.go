package response

import (
	"go-apevolo/model/permission"
)

type Login struct {
	User  JwtUser  `json:"user"`
	Token JwtToken `json:"token"`
}

type JwtUser struct {
	User      permission.User `json:"user"`
	Role      []string        `json:"roles"`
	DataScope []string        `json:"dataScopes"`
}

type JwtToken struct {
	AccessToken         string `json:"access_token"`
	Expires             int32  `json:"expires_in"`
	TokenType           string `json:"token_type"`
	RefreshToken        string `json:"refresh_token"`
	RefreshTokenExpires int32  `json:"refresh_token_expires_in"`
}

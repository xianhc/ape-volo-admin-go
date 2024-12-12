package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"go-apevolo/global"
	"go-apevolo/payloads/request"
	"time"
)

type JwtAuthOptions struct {
	SigningKey []byte
}

//var (
//	TokenExpired     = errors.New("Token is expired.")
//	TokenNotValidYet = errors.New("Token not active yet")
//	TokenMalformed   = errors.New("That's not even a token")
//	TokenInvalid     = errors.New("Couldn't handle this token:")
//)

func NewJwt() *JwtAuthOptions {
	return &JwtAuthOptions{
		[]byte(global.Config.JwtAuthOptions.SecurityKey),
	}
}

// IssuedToken 创建token
func (j *JwtAuthOptions) IssuedToken(claims request.Claims) (string, error) {
	currentTime := time.Now()
	expiresAt := currentTime.Add(time.Duration(global.Config.JwtAuthOptions.Expires) * time.Hour)
	apeClaims := request.ApeClaims{
		Claims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    global.Config.JwtAuthOptions.Issuer,
			Audience:  jwt.ClaimStrings{global.Config.JwtAuthOptions.Audience},
			NotBefore: jwt.NewNumericDate(currentTime.Add(-1000)),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, apeClaims)
	return token.SignedString(j.SigningKey)
}

// ReadJwtToken 解析token
func (j *JwtAuthOptions) ReadJwtToken(tokenString string) (*request.ApeClaims, error) {
	claims := &request.ApeClaims{}
	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 检查 Token 是否有效
	if token != nil && token.Valid {
		if claims, ok := token.Claims.(*request.ApeClaims); ok {
			return claims, nil
		}
		return nil, jwt.ErrInvalidKey
	}

	return nil, jwt.ErrTokenUnverifiable
}

//func (j *JwtAuthOptions) ReadJwtToken(tokenString string) (*request.ApeClaims, error) {
//	token, err := jwt.ParseWithClaims(tokenString, &request.ApeClaims{}, func(token *jwt.Token) (i interface{}, e error) {
//		return j.SigningKey, nil
//	})
//	if err != nil {
//		var ve *jwt.ValidationError
//		if errors.As(err, &ve) {
//			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
//				return nil, TokenMalformed
//			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
//				return nil, TokenExpired
//			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
//				return nil, TokenNotValidYet
//			} else {
//				return nil, TokenInvalid
//			}
//		}
//	}
//	if token != nil {
//		if claims, ok := token.Claims.(*request.ApeClaims); ok && token.Valid {
//			return claims, nil
//		}
//		return nil, TokenInvalid
//
//	} else {
//		return nil, TokenInvalid
//	}
//}

package utils

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	jwt "go-apevolo/payloads/request"
	"go-apevolo/utils/ext"
)

func GetClaims(c *gin.Context) (*jwt.ApeClaims, error) {
	token := GetToken(c)
	token = ext.StringReplace(token, "Bearer ", "", -1)
	j := NewJwt()
	claims, err := j.ReadJwtToken(token)
	if err != nil {
		global.Logger.Error("Authorization Token解析失败")
	}
	return claims, err
}

func GetId(c *gin.Context) int64 {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0
		} else {
			return cl.Claims.Jti
		}
	} else {
		waitUse := claims.(*jwt.ApeClaims)
		return waitUse.Claims.Jti
	}
}

func GetAccount(c *gin.Context) string {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return ""
		} else {
			return cl.Claims.Name
		}
	} else {
		waitUse := claims.(*jwt.ApeClaims)
		return waitUse.Claims.Name
	}
}

func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	return token
}

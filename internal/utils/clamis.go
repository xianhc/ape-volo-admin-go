package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

func GetClaims(c *gin.Context) (*ApeClaims, error) {
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
			return UserID(cl)
		}
	} else {
		waitUse := claims.(*ApeClaims)
		return UserID(waitUse)
	}
}

func GetAccount(c *gin.Context) string {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return ""
		} else {
			return cl.Name
		}
	} else {
		waitUse := claims.(*ApeClaims)
		return waitUse.Name
	}
}

func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	return token
}

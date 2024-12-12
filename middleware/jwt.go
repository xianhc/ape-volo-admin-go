package middleware

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetToken(c)
		if token == "" {
			response.Error("抱歉，您无权访问该接口", nil, c)
			c.Abort()
			return
		}
		token = ext.StringReplace(token, "Bearer ", "", -1)
		j := utils.NewJwt()
		_, err := j.ReadJwtToken(token)
		if err != nil {
			response.Error(err.Error(), nil, c)
			c.Abort()
			return
		}
		c.Next()
	}
}

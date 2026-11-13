package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

// JWTAuth 校验 Authorization 头中的 token。
//
// 未携带 token 或 token 无效时返回 401（此前返回 400），
// 使前端能区分「未登录」与「参数错误」。
// 解析出的 claims 写入上下文，后续处理无需重复解析。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ext.StringReplace(utils.GetToken(c), "Bearer ", "", -1)
		if token == "" {
			response.Unauthorized("未提供认证信息，请先登录", nil, c)
			c.Abort()
			return
		}

		claims, err := utils.NewJwt().ReadJwtToken(token)
		if err != nil {
			response.Unauthorized("登录状态已失效，请重新登录", nil, c)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		// 同时写入请求上下文：ORM 写回调只拿得到 Request.Context()，
		// 审计字段的自动填充依赖它。
		c.Request = c.Request.WithContext(utils.WithOperatorID(c.Request.Context(), utils.UserID(claims)))
		c.Next()
	}
}

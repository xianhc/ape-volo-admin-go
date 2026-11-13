package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// BindParams 绑定查询参数：GET / DELETE 走 query，POST / PUT 优先 JSON body、没有 body 时回落到 query。
//
// 为什么需要它：同一个接口在不同调用方眼里传参方式不同——
// 导出接口前端用 POST + JSON body（`data: params`），而字典、角色详情的 POST 却把参数放在 query 上
// （axios 的 `params`）。只按一种方式绑会丢掉筛选条件，导出的就是全量数据。
func BindParams(c *gin.Context, obj any) error {
	switch c.Request.Method {
	case http.MethodGet, http.MethodDelete:
		return c.ShouldBindQuery(obj)
	}

	if c.Request.ContentLength > 0 {
		return c.ShouldBindJSON(obj)
	}
	return c.ShouldBindQuery(obj)
}

// BindAndValidateJSON 绑定并校验 JSON 请求，失败时写入参数错误响应。
//
// 返回 false 时 Handler 应立即返回；校验错误仍保留按字段展示的响应结构。
func BindAndValidateJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.Error(err.Error(), nil, c)
		return false
	}
	if err := utils.VerifyData(obj); err != nil {
		response.Error("", response.GetVerifyErr(err), c)
		return false
	}
	return true
}

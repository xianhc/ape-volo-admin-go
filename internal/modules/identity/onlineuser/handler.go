package onlineuser

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
)

// Handler 处理在线用户查询、强制下线和导出请求。
//
// Handler 只负责参数绑定、调用 Service 和写入响应。
type Handler struct{ service Service }

func requireRedis(c *gin.Context) bool {
	if global.Config.System.UseRedis {
		return true
	}
	response.Error("在线用户功能仅支持 Redis，请启用 system.use-redis。", nil, c)
	return false
}

// Query 查询在线用户分页列表。
//
// 成功时返回统一分页结构。
// @Tags OnlineUser
// @Router /online/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	if !requireRedis(c) {
		return
	}
	var req QueryRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	items, total, err := handler.service.Query(c.Request.Context(), req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// DropOut 强制撤销一组在线会话。
//
// 请求主键实际为会话标识。
// @Tags OnlineUser
// @Router /online/out [delete]
func (handler *Handler) DropOut(c *gin.Context) {
	if !requireRedis(c) {
		return
	}
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.DropOut(c.Request.Context(), ids.IdArray); err != nil {
		response.Error("登出失败:"+err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Download 导出在线用户列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags OnlineUser
// @Router /online/download [get]
func (handler *Handler) Download(c *gin.Context) {
	if !requireRedis(c) {
		return
	}
	path, name, err := handler.service.Download(c.Request.Context())
	if err != nil {
		response.Error("导出失败", nil, c)
		return
	}
	response.DownloadFile(c, path, name)
}

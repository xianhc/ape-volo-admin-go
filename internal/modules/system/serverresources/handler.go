package serverresources

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 处理服务器资源 HTTP 请求。
//
// Handler 只调用 Service 并写入 JSON 响应。
type Handler struct {
	service Service
}

// Query 查询当前服务器资源快照。
//
// 成功时直接返回 ServerResourcesResponse，保持既有 HTTP 200 契约。
// @Tags ServerResourcesInfo
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Success 200 {object} serverresources.ServerResourcesResponse "查询成功"
// @Router /server-resources/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	c.JSON(http.StatusOK, handler.service.Query())
}

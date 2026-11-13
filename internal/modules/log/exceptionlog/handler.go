package exceptionlog

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
)

// Handler 处理异常日志的 HTTP 请求。
//
// Handler 只负责参数绑定、调用 Service 和写入统一分页响应。
type Handler struct{ service Service }

// Query 查询异常日志分页列表。
//
// 查询参数同时兼容 URL query 与请求体。
// @Tags ExceptionLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body exceptionlog.QueryExceptionLogRequest true "log request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /exception-log/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryExceptionLogRequest
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

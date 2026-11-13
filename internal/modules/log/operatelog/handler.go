package operatelog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理操作日志的 HTTP 请求。
//
// Handler 只负责参数绑定、认证上下文读取、调用 Service 和写入响应。
type Handler struct{ service Service }

// Query 查询操作日志分页列表。
//
// 查询参数同时兼容 URL query 与请求体。
// @Tags AuditLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param data body operatelog.QueryOperateLogRequest false "查询参数"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /operate-log/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryOperateLogRequest
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

// GetByCurrent 查询当前用户的操作日志。
//
// 当前用户主键从 JWT 上下文读取，客户端不能覆盖。
// @Tags AuditLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body operatelog.QueryCurrentRequest false "log request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /operate-log/get-by-current [get]
func (handler *Handler) GetByCurrent(c *gin.Context) {
	var req QueryCurrentRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	items, total, err := handler.service.QueryByCurrent(c.Request.Context(), req, utils.GetId(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// GetOperationNumber 统计最近若干天的每日操作数。
//
// days 缺失或无法解析时使用 7，结果直接返回整数数组。
// @Tags AuditLog
// @Summary 统计最近 N 天每天的操作数
// @Security ApiKeyAuth
// @Produce application/json
// @Param days query int false "天数，默认 7"
// @Success 200 {object} []int "每天的操作数"
// @Router /operate-log/get-operation-number [get]
func (handler *Handler) GetOperationNumber(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	numbers, err := handler.service.GetOperationNumber(c.Request.Context(), days)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, numbers)
}

package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/monitor"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type AuditLogApi struct{}

// Query
// @Tags   AuditLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param  data body dto.LogQueryCriteria false  "查询参数"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /api/auditing/query [get]
func (a *AuditLogApi) Query(c *gin.Context) {
	var pageInfo dto.LogQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	list := make([]monitor.AuditLog, 0)
	var count int64
	err = auditLogService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

// QueryByCurrent
// @Tags   AuditLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body dto.LogQueryCriteria false "log request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /api/auditing/current [get]
func (a *AuditLogApi) QueryByCurrent(c *gin.Context) {
	var pageInfo request.Pagination
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	list := make([]monitor.AuditLog, 0)
	var count int64
	err = auditLogService.QueryByCurrent(pageInfo, utils.GetAccount(c), &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/monitor"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/response"
)

type ExceptionLogApi struct{}

// Query
// @Tags   ExceptionLog
// @Summary 查询
// @Accept application/json
// @Produce application/json
// @Param request body dto.LogQueryCriteria true "log request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /exception/query [get]
func (e *ExceptionLogApi) Query(c *gin.Context) {
	var pageInfo dto.LogQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	list := make([]monitor.ExceptionLog, 0)
	var count int64
	err = exceptionLogService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

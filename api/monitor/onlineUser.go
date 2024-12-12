package monitor

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go.uber.org/zap"
)

type OnlineUserApi struct{}

// Query
// @Tags   OnlineUser
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body request.Pagination false " "
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /api/online/query [get]
func (o *OnlineUserApi) Query(c *gin.Context) {
	var pageInfo request.Pagination
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	list, total, err := onlineUserService.Query(pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: total,
	}, c)
}

// DropOut
// @Tags   OnlineUser
// @Summary 用户登出
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID集合"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /api/online/out [delete]
func (o *OnlineUserApi) DropOut(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = onlineUserService.DropOut(idArray)
	if err != nil {
		global.Logger.Error("登出失败!", zap.Error(err))
		response.Error("登出失败:"+err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go.uber.org/zap"
)

type ApisApi struct{}

// Create
// @Tags   Apis
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateApisDto true "CreateUpdateApisDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /apis/create [post]
func (a *ApisApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateApisDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	reqInfo.SetCreateBy(utils.GetAccount(c))
	err = apisService.Create(&reqInfo)
	if err != nil {
		global.Logger.Error("创建失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Apis
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateApisDto true "CreateUpdateApisDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /apis/edit [put]
func (a *ApisApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateApisDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	reqInfo.SetUpdateBy(utils.GetAccount(c))
	err = apisService.Update(&reqInfo)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Apis
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /apis/delete [delete]
func (a *ApisApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = apisService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		global.Logger.Error("删除失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Apis
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.ApisQueryCriteria true "Apis request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /apis/query [get]
func (a *ApisApi) Query(c *gin.Context) {
	var pageInfo dto.ApisQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(pageInfo.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	list := make([]permission.Apis, 0)
	var count int64
	err = apisService.Query(pageInfo, &list, &count)
	if err != nil {
		response.Error("获取失败", nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

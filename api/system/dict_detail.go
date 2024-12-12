package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
)

type DictDetailApi struct{}

// Create
// @Tags   DictDetail
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDictDetailDto true "CreateUpdateDictDetailDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /dictDetail/create [post]
func (d *DictDetailApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateDictDetailDto
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
	err = dictDetailService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   DictDetail
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDictDetailDto true "CreateUpdateDictDetailDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dictDetail/edit [put]
func (d *DictDetailApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateDictDetailDto
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
	err = dictDetailService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Dict
// @Summary 删除
// @Accept json
// @Produce json
// @Param id query int64 true "id"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /dictDetail/delete [delete]
func (d *DictDetailApi) Delete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	err := dictDetailService.Delete(ext.StringToInt64(id), utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   DictDetail
// @Summary 查询
// @Accept json
// @Produce json
// @Param dictName query string true "dictName"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dictDetail/query [get]
func (d *DictDetailApi) Query(c *gin.Context) {
	dictName := c.Query("dictName")
	if dictName == "" {
		response.Error("dictName is null", nil, c)
		return
	}
	list := make([]system.DictDetail, 0)
	err := dictDetailService.Query(dictName, &list)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: int64(len(list)),
	}, c)
}

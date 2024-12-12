package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type DictApi struct{}

// Create
// @Tags   Dict
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDictDto true "CreateUpdateDictDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /dict/create [post]
func (d *DictApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateDictDto
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
	err = dictService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Dict
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateDictDto true "CreateUpdateDictDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dict/edit [put]
func (d *DictApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateDictDto
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
	err = dictService.Update(&reqInfo)
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
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /dict/edit [delete]
func (d *DictApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = dictService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Dict
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.DictQueryCriteria true "Dict request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dict/query [get]
func (d *DictApi) Query(c *gin.Context) {
	var pageInfo dto.DictQueryCriteria
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
	list := make([]system.Dict, 0)
	var count int64
	err = dictService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

// Download
// @Tags   Dict
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.DictQueryCriteria true "Dict request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dict/download [get]
func (d *DictApi) Download(c *gin.Context) {
	var pageInfo dto.DictQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := dictService.Download(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	// 设置响应头，告诉浏览器将文件作为附件下载
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.File(filePath)
}

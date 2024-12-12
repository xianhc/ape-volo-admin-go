package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type AppSecretApi struct{}

// Create
// @Tags   AppSecret
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateAppSecretDto true "CreateUpdateAppSecretDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /appSecret/create [post]
func (a *AppSecretApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateAppSecretDto
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
	err = appSecretService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   AppSecret
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateAppSecretDto true "CreateUpdateAppSecretDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /appSecret/edit [put]
func (a *AppSecretApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateAppSecretDto
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
	err = appSecretService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   AppSecret
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /appSecret/edit [delete]
func (a *AppSecretApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = appSecretService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   AppSecret
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.AppSecretQueryCriteria true "AppSecret request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /appSecret/query [get]
func (a *AppSecretApi) Query(c *gin.Context) {
	var pageInfo dto.AppSecretQueryCriteria
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
	list := make([]system.AppSecret, 0)
	var count int64
	err = appSecretService.Query(&pageInfo, &list, &count)
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
// @Tags   AppSecret
// @Summary 导出
// @Accept json
// @Produce json
// @Param request body dto.AppSecretQueryCriteria true "AppSecret request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /appSecret/download [get]
func (a *AppSecretApi) Download(c *gin.Context) {
	var pageInfo dto.AppSecretQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := appSecretService.Download(&pageInfo)
	if err != nil {
		response.Error("导出失败", nil, c)
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

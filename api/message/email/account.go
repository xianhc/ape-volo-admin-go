package email

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type AccountApi struct{}

// Create
// @Tags   EmailAccount
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailAccountDto true "CreateUpdateEmailAccountDto"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /email/account/create [post]
func (a *AccountApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailAccountDto
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
	err = emailAccountService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   EmailAccount
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailAccountDto true "CreateUpdateEmailAccountDto"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /email/account/edit [put]
func (a *AccountApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailAccountDto
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
	err = emailAccountService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   EmailAccount
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /email/account/delete [delete]
func (a *AccountApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = emailAccountService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags EmailAccount
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.EmailAccountQueryCriteria true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /email/account/query [get]
func (a *AccountApi) Query(c *gin.Context) {
	var reqInfo dto.EmailAccountQueryCriteria
	err := c.ShouldBindQuery(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	list := make([]email.Account, 0)
	var count int64
	err = emailAccountService.Query(&reqInfo, &list, &count)
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
// @Tags   EmailAccount
// @Summary 导出
// @Accept json
// @Produce json
// @Param request body dto.EmailAccountQueryCriteria true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /email/account/download [get]
func (a *AccountApi) Download(c *gin.Context) {
	var emailAccountQueryCriteria dto.EmailAccountQueryCriteria
	err := c.ShouldBindQuery(&emailAccountQueryCriteria)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := emailAccountService.Download(&emailAccountQueryCriteria)
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

package email

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/message/email"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type MessageTemplateApi struct{}

// Create
// @Tags   EmailMessageTemplate
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailMessageTemplateDto true "CreateUpdateEmailMessageTemplateDto"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /email/template/create [post]
func (a *MessageTemplateApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailMessageTemplateDto
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
	err = emailMessageTemplateService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   EmailMessageTemplate
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailMessageTemplateDto true "CreateUpdateEmailMessageTemplateDto"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /email/template/edit [put]
func (a *MessageTemplateApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateEmailMessageTemplateDto
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
	err = emailMessageTemplateService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   EmailMessageTemplate
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /email/template/delete [delete]
func (a *MessageTemplateApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = emailMessageTemplateService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags EmailMessageTemplate
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateEmailMessageTemplateDto true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /email/template/query [get]
func (a *MessageTemplateApi) Query(c *gin.Context) {
	var reqInfo dto.EmailMessageTemplateQueryCriteria
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
	list := make([]email.MessageTemplate, 0)
	var count int64
	err = emailMessageTemplateService.Query(&reqInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

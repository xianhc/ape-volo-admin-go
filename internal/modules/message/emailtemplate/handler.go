package emailtemplate

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理邮件模板的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建邮件模板。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags EmailMessageTemplate
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body emailtemplate.CreateAndUpdateTemplateRequest true "CreateAndUpdateTemplateRequest"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /email-message-template/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateTemplateRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改邮件模板。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 204。
// @Tags EmailMessageTemplate
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body emailtemplate.CreateAndUpdateTemplateRequest true "CreateAndUpdateTemplateRequest"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /email-message-template/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateTemplateRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组邮件模板。
//
// 主键集合兼容对象和裸数组两种已有请求格式。
// @Tags EmailMessageTemplate
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /email-message-template/delete [delete]
func (handler *Handler) Delete(c *gin.Context) {
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.Delete(c.Request.Context(), ids.IdArray); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 查询邮件模板分页列表。
//
// 查询支持名称、激活状态和创建时间范围。
// @Tags EmailMessageTemplate
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body emailtemplate.QueryTemplateRequest true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /email-message-template/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryTemplateRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := utils.VerifyData(req); err != nil {
		response.Error("", response.GetVerifyErr(err), c)
		return
	}
	items, total, err := handler.service.Query(c.Request.Context(), req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

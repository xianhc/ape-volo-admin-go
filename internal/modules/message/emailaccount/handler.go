package emailaccount

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理邮件账户的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建邮件账户。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags EmailAccount
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body emailaccount.CreateAndUpdateAccountRequest true "CreateAndUpdateAccountRequest"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /email-account/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateAccountRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改邮件账户。
//
// 密码留空表示不修改，成功时返回 HTTP 204。
// @Tags EmailAccount
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body emailaccount.CreateAndUpdateAccountRequest true "CreateAndUpdateAccountRequest"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /email-account/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateAccountRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组邮件账户。
//
// 主键集合兼容对象和裸数组两种已有请求格式。
// @Tags EmailAccount
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /email-account/delete [delete]
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

// Query 查询邮件账户分页列表。
//
// 返回模型中的密码字段始终为空。
// @Tags EmailAccount
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body emailaccount.QueryAccountRequest true "request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /email-account/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryAccountRequest
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

// All 查询全部邮件账户。
//
// 该接口直接返回脱敏后的数组，供前端下拉框使用。
// @Tags EmailAccount
// @Summary 查询全部邮箱账户
// @Accept json
// @Produce json
// @Success 200 {object} []emailaccount.AccountResponse "账户列表"
// @Router /email-account/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.All(c.Request.Context())
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

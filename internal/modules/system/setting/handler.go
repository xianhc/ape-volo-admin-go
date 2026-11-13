package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理系统设置的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入统一响应。
type Handler struct {
	service Service
}

// Create 创建系统设置。
//
// 请求校验通过后交给 Service 处理，成功时返回 HTTP 201。
// @Tags   Setting
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body setting.CreateAndUpdateSettingRequest true "CreateAndUpdateSettingRequest object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /setting/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateSettingRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改系统设置。
//
// 请求校验通过后交给 Service 处理，成功时返回 HTTP 204。
// @Tags   Setting
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body setting.CreateAndUpdateSettingRequest true "CreateAndUpdateSettingRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /setting/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateSettingRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组系统设置。
//
// 主键集合兼容对象和裸数组两种已有请求格式。
// @Tags   Setting
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /setting/delete [delete]
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

// Query 查询系统设置分页列表。
//
// 查询参数同时兼容 URL query 与请求体，成功时返回统一分页结构。
// @Tags   Setting
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body setting.QuerySettingRequest true "Setting request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /setting/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QuerySettingRequest
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

// Download 导出系统设置列表。
//
// 导出成功后以 XLSX 附件响应，筛选条件与列表查询保持一致。
// @Tags   Setting
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body setting.QuerySettingRequest true "Setting request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /setting/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QuerySettingRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := handler.service.Download(c.Request.Context(), req)
	if err != nil {
		response.Error("导出失败", nil, c)
		return
	}
	response.DownloadFile(c, filePath, fileName)
}

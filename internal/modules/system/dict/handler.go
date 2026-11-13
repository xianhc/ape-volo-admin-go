package dict

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

// DictHandler 处理字典主表的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type DictHandler struct{ service Service }

// Create 创建字典主表。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags Dict
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dict.CreateAndUpdateDictRequest true "CreateAndUpdateDictRequest object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /dict/create [post]
func (handler *DictHandler) Create(c *gin.Context) {
	var req CreateAndUpdateDictRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.CreateDict(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改字典主表。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 204。
// @Tags Dict
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dict.CreateAndUpdateDictRequest true "CreateAndUpdateDictRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dict/update [put]
func (handler *DictHandler) Update(c *gin.Context) {
	var req CreateAndUpdateDictRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdateDict(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组字典主表。
//
// 主键集合兼容对象和裸数组两种已有请求格式。
// @Tags Dict
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /dict/delete [delete]
func (handler *DictHandler) Delete(c *gin.Context) {
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.DeleteDicts(c.Request.Context(), ids.IdArray); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 查询字典分页列表。
//
// 成功时返回包含字典明细的统一分页结构。
// @Tags Dict
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dict.QueryDictRequest true "Dict request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dict/get-list [get]
func (handler *DictHandler) Query(c *gin.Context) {
	var req QueryDictRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := utils.VerifyData(req); err != nil {
		response.Error("", response.GetVerifyErr(err), c)
		return
	}
	items, total, err := handler.service.QueryDicts(c.Request.Context(), req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出字典及其明细。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Dict
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dict.QueryDictRequest true "Dict request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dict/download [post]
func (handler *DictHandler) Download(c *gin.Context) {
	var req QueryDictRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	path, name, err := handler.service.DownloadDicts(c.Request.Context(), req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	setFileHeaders(c, name)
	c.File(path)
}

// GetByName 按名称查询字典及其明细。
//
// 该接口直接返回独立 DictResponse，不使用分页响应壳。
// @Tags Dict
// @Summary 按字典名称查询
// @Security ApiKeyAuth
// @Produce application/json
// @Param name query string true "字典名称"
// @Success 200 {object} dict.DictResponse "字典及其详情"
// @Router /dict/get-by-name [get]
func (handler *DictHandler) GetByName(c *gin.Context) {
	dict, err := handler.service.GetDictByName(c.Request.Context(), c.Query("name"))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, dict)
}

// DetailHandler 处理字典明细的 HTTP 请求。
//
// Handler 与 DictHandler 共用同一个模块 Service。
type DetailHandler struct{ service Service }

// Create 创建字典明细。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags DictDetail
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dict.CreateAndUpdateDictDetailRequest true "CreateAndUpdateDictDetailRequest object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /dict-detail/create [post]
func (handler *DetailHandler) Create(c *gin.Context) {
	var req CreateAndUpdateDictDetailRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.CreateDetail(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改字典明细。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 204。
// @Tags DictDetail
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dict.CreateAndUpdateDictDetailRequest true "CreateAndUpdateDictDetailRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /dict-detail/update [put]
func (handler *DetailHandler) Update(c *gin.Context) {
	var req CreateAndUpdateDictDetailRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdateDetail(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除字典明细。
//
// 主键继续从 query 参数读取，保持既有接口契约。
// @Tags Dict
// @Summary 删除
// @Accept json
// @Produce json
// @Param id query int64 true "id"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /dict-detail/delete [delete]
func (handler *DetailHandler) Delete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.DeleteDetail(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 按字典名称查询全部明细。
//
// 成功时沿用统一分页形状，总数等于返回明细数量。
// @Tags DictDetail
// @Summary 查询
// @Accept json
// @Produce json
// @Param dictName query string true "dictName"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /dict-detail/get-list [get]
func (handler *DetailHandler) Query(c *gin.Context) {
	name := c.Query("dictName")
	if name == "" {
		response.Error("dictName is null", nil, c)
		return
	}
	items, err := handler.service.QueryDetails(c.Request.Context(), name)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: int64(len(items))}, c)
}

func setFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
}

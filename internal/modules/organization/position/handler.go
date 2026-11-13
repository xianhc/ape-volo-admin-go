package position

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理岗位的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建岗位。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags Position
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body position.CreateAndUpdatePositionRequest true "CreateAndUpdatePositionRequest object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /position/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdatePositionRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改岗位。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 204。
// @Tags Position
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body position.CreateAndUpdatePositionRequest true "CreateAndUpdatePositionRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /position/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdatePositionRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组岗位。
//
// 被用户引用的岗位不能删除。
// @Tags Position
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /position/delete [delete]
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

// Query 查询岗位分页列表。
//
// 查询参数同时兼容 URL query 与请求体。
// @Tags Position
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body position.QueryPositionRequest true "Job request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /position/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryPositionRequest
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
		response.Error("获取失败", nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出岗位列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Position
// @Router /position/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryPositionRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	path, name, err := handler.service.Download(c.Request.Context(), req)
	if err != nil {
		response.Error("导出失败", nil, c)
		return
	}
	response.DownloadFile(c, path, name)
}

// All 查询全部启用岗位。
//
// 该接口直接返回按排序字段升序排列的 Response 数组。
// @Tags Position
// @Summary 查询全部岗位
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /position/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.All(c.Request.Context())
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

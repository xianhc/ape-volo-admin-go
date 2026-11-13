package department

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
)

// Handler 处理部门的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建部门。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 201。
// @Tags Dept
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body department.CreateAndUpdateDepartmentRequest true "CreateAndUpdateDepartmentRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /department/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateDepartmentRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改部门。
//
// 请求校验通过后交给 Service，成功时返回 HTTP 204。
// @Tags Dept
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body department.CreateAndUpdateDepartmentRequest true "CreateAndUpdateDepartmentRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /department/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateDepartmentRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除部门及其子部门。
//
// 字符串主键集合转换为 int64 后交给 Service。
// @Tags Dept
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /department/delete [delete]
func (handler *Handler) Delete(c *gin.Context) {
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	values := make([]int64, 0, len(ids.IdArray))
	for _, raw := range ids.IdArray {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			values = append(values, value)
		}
	}
	if err := handler.service.Delete(c.Request.Context(), values); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 查询部门列表。
//
// 根节点查询分页，指定父节点查询全部直接子节点。
// @Tags Dept
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body department.QueryDepartmentRequest true "Dept request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /department/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryDepartmentRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	items, total, err := handler.service.Query(c.Request.Context(), req)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出部门列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Dept
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body department.QueryDepartmentRequest true "Dept request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /department/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryDepartmentRequest
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

// All 查询全部部门。
//
// 该接口直接返回按 Sort 排序的 Response 数组。
// @Tags Dept
// @Summary 查询全部部门
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} []department.DepartmentResponse "部门列表"
// @Router /department/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.All(c.Request.Context())
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

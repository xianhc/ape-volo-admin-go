package role

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

// Handler 处理角色管理与菜单授权的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建角色。
//
// 成功时返回 HTTP 201。
// @Tags Role
// @Router /role/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateRoleRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改角色。
//
// 成功时返回 HTTP 204。
// @Tags Role
// @Router /role/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateRoleRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组角色。
//
// 被用户引用的角色不能删除。
// @Tags Role
// @Router /role/delete [delete]
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

// Query 查询角色分页列表。
//
// 返回结果包含菜单关联。
// @Tags Role
// @Router /role/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryRoleRequest
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

// Download 导出角色列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Role
// @Router /role/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryRoleRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	path, name, err := handler.service.Download(c.Request.Context(), req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.DownloadFile(c, path, name)
}

// All 查询全部角色。
//
// 该接口直接返回包含关联的 Response 数组。
// @Tags Role
// @Router /role/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.All(c.Request.Context())
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetWithDetails 查询单个角色及完整关联。
//
// 主键继续从 query 参数读取。
// @Tags Role
// @Router /role/get-with-details [get]
func (handler *Handler) GetWithDetails(c *gin.Context) {
	raw := c.Query("id")
	if raw == "" {
		response.Error("id is null", nil, c)
		return
	}
	item, err := handler.service.GetWithDetails(c.Request.Context(), ext.StringToInt64(raw))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, item)
}

// UpdateMenus 替换角色菜单授权。
//
// 成功时返回 HTTP 204。
// @Tags Role
// @Router /role/update-role-menu [put]
func (handler *Handler) UpdateMenus(c *gin.Context) {
	var req CreateAndUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.UpdateMenus(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

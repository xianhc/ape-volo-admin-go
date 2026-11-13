package menu

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理菜单配置、树查询和用户路由的 HTTP 请求。
//
// Handler 只负责参数绑定、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建菜单。
//
// 业务组合校验由 Service 处理。
// @Tags Menu
// @Router /menu/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateMenuRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改菜单。
//
// 成功时返回 HTTP 204。
// @Tags Menu
// @Router /menu/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateMenuRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除菜单及其子菜单。
//
// 主键集合兼容对象和裸数组请求格式。
// @Tags Menu
// @Router /menu/delete [delete]
func (handler *Handler) Delete(c *gin.Context) {
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	values := make([]int64, 0, len(ids.IdArray))
	for _, raw := range ids.IdArray {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
			values = append(values, id)
		}
	}
	if err := handler.service.Delete(c.Request.Context(), values); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 查询直接菜单列表。
//
// 成功时返回统一分页形状。
// @Tags Menu
// @Router /menu/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryMenuRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	items, total, err := handler.service.Query(c.Request.Context(), req)
	if err != nil {
		response.Error("获取失败", nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出菜单列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Menu
// @Router /menu/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryMenuRequest
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

// Build 构建当前用户前端路由菜单。
//
// 用户主键从认证上下文读取，结果由 Service 缓存。
// @Tags Menu
// @Router /menu/build-tree [post]
func (handler *Handler) Build(c *gin.Context) {
	items, err := handler.service.BuildTree(c.Request.Context(), utils.GetId(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

// All 查询完整菜单授权树。
//
// 该接口供角色菜单授权弹窗使用。
// @Tags Menu
// @Router /menu/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.AllTree(c.Request.Context())
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

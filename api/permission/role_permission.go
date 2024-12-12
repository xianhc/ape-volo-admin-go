package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"net/http"
)

type RolePermissionApi struct{}

// QueryAllMenus
// @Tags   RolePermissionApi
// @Summary 查询
// @Accept json
// @Produce json
// @Success 200 {object} []permission.Menu "查询成功"
// @Router /api/permissions/menus/query [get]
func (a *RolePermissionApi) QueryAllMenus(c *gin.Context) {
	list, err := rolePermissionService.GetAllMenus()
	if err != nil {
		response.Error("获取菜单Tree失败", nil, c)
		return
	}
	c.JSON(
		http.StatusOK,
		list,
	)
}

// QueryAllApis
// @Tags   RolePermissionApi
// @Summary 查询
// @Accept json
// @Produce json
// @Success 200 {object} []vo.ApisTree "查询成功"
// @Router /api/permissions/apis/query [get]
func (a *RolePermissionApi) QueryAllApis(c *gin.Context) {
	list, err := rolePermissionService.GetAllApis()
	if err != nil {
		response.Error("获取ApisTree失败", nil, c)
		return
	}
	c.JSON(
		http.StatusOK,
		list,
	)
}

// UpdateRolesMenus
// @Tags   RolePermissionApi
// @Summary 更新角色菜单关联
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateRoleDto true "CreateUpdateJobDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /api/permissions/menus/edit [get]
func (a *RolePermissionApi) UpdateRolesMenus(c *gin.Context) {
	var reqInfo dto.CreateUpdateRoleDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = rolePermissionService.UpdateRolesMenus(&reqInfo, utils.GetId(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdateRolesApis
// @Tags   RolePermissionApi
// @Summary 更新角色Apis关联
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateRoleDto true "CreateUpdateJobDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /api/permissions/apis/edit [get]
func (a *RolePermissionApi) UpdateRolesApis(c *gin.Context) {
	var reqInfo dto.CreateUpdateRoleDto
	err := c.ShouldBindJSON(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = rolePermissionService.UpdateRolesApis(&reqInfo, utils.GetId(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

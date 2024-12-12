package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"net/http"
)

type RoleApi struct{}

// Create
// @Tags   Role
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateRoleDto true "CreateUpdateRoleDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /role/create [post]
func (r *RoleApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateRoleDto
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
	err = roleService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Role
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateRoleDto true "CreateUpdateJobDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /role/edit [put]
func (r *RoleApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateRoleDto
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
	err = roleService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Role
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /role/delete [delete]
func (r *RoleApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = roleService.Delete(idArray, utils.GetId(c), utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Role
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.RoleQueryCriteria true "Role request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /role/query [get]
func (r *RoleApi) Query(c *gin.Context) {
	var pageInfo dto.RoleQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(pageInfo.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	list := make([]permission.Role, 0)
	var count int64
	err = roleService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

// Download
// @Tags   Role
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.RoleQueryCriteria true "Role request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /role/download [get]
func (r *RoleApi) Download(c *gin.Context) {
	var pageInfo dto.RoleQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := roleService.Download(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	// 设置响应头，告诉浏览器将文件作为附件下载
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.File(filePath)
}

// All
// @Tags   Role
// @Summary 查询全部角色
// @Accept json
// @Produce json
// @Success 200 {object} []permission.Role "查询成功"
// @Router /role/all [get]
func (r *RoleApi) All(c *gin.Context) {
	list := make([]permission.Role, 0)
	err := roleService.GetAllRole(&list)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	for i := range list {
		for j := range list[i].Menus {
			list[i].Menus[j].CalculateHasChildren()
			list[i].Menus[j].CalculateLeaf()
			list[i].Menus[j].InitChildren()
		}
	}
	c.JSON(http.StatusOK, list)
}

// Level
// @Tags   Role
// @Summary 查询
// @Accept json
// @Produce json
// @Success 200 {object} response.RoleLevel "查询成功"
// @Router /role/level [get]
func (r *RoleApi) Level(c *gin.Context) {
	userId := utils.GetId(c)
	level, err := roleService.VerificationUserRoleLevelAsync(userId, nil)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, response.RoleLevel{
		Level: level,
	})
}

// QuerySingle
// @Tags   Role
// @Summary 查看单一角色
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /role/querySingle [get]
func (r *RoleApi) QuerySingle(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	role := permission.Role{}
	err := roleService.QuerySingle(ext.StringToInt64(id), &role)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, role)
}

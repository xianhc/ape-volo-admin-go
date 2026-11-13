package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理用户管理和个人中心 HTTP 请求。
//
// Handler 只负责参数绑定、校验、认证用户读取、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建用户。
//
// 成功时返回 HTTP 201。
// @Tags User
// @Router /user/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateUserRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改用户。
//
// 成功时返回 HTTP 204。
// @Tags User
// @Router /user/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateUserRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组用户。
//
// 当前用户主键从认证上下文读取，Service 禁止自删。
// @Tags User
// @Router /user/delete [delete]
func (handler *Handler) Delete(c *gin.Context) {
	var ids request.IdCollection
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := handler.service.Delete(c.Request.Context(), ids.IdArray, utils.GetId(c)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query 查询用户分页列表。
//
// 返回脱敏 UserResponse 与统一分页结构。
// @Tags User
// @Summary 查询用户列表
// @Param departmentIdArray query string false "逗号分隔的部门主键，仅匹配直属用户"
// @Param id query string false "用户主键"
// @Param userName query string false "用户名，兼容 username"
// @Param nickName query string false "昵称，兼容 nickname"
// @Param enabled query bool false "是否启用"
// @Param pageIndex query int true "页码"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} response.ActionResultPage
// @Router /user/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryUserRequest
	if err := bindQuery(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if err := utils.VerifyData(req); err != nil {
		response.Error("", response.GetVerifyErr(err), c)
		return
	}
	items, total, err := handler.service.Query(c.Request.Context(), req)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出用户列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags User
// @Router /user/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryUserRequest
	if err := bindQuery(c, &req); err != nil {
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

// bindQuery 同时接受驼峰及全小写的用户名、昵称查询字段。
func bindQuery(c *gin.Context, req *QueryUserRequest) error {
	if err := request.BindParams(c, req); err != nil {
		return err
	}
	if req.UserName == "" {
		req.UserName = c.Query("username")
	}
	if req.NickName == "" {
		req.NickName = c.Query("nickname")
	}
	return nil
}

// UpdateCenter 修改当前用户基础信息。
//
// 用户主键从认证上下文读取。
// @Tags UserCenter
// @Router /user/update-center [put]
func (handler *Handler) UpdateCenter(c *gin.Context) {
	var req UpdateCenterRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdateCenter(c.Request.Context(), utils.GetId(c), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdatePassword 修改当前用户密码。
//
// 成功后 Service 撤销该用户全部会话。
// @Tags UserCenter
// @Router /user/update-password [put]
func (handler *Handler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdatePassword(c.Request.Context(), utils.GetId(c), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdateEmail 修改当前用户邮箱。
//
// Service 校验登录密码和一次性邮箱验证码。
// @Tags UserCenter
// @Router /user/update-email [put]
func (handler *Handler) UpdateEmail(c *gin.Context) {
	var req UpdateEmailRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdateEmail(c.Request.Context(), utils.GetId(c), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdatePositions 替换请求指定用户的岗位关联。
//
// 岗位主键集合通过请求体验证。
// @Tags UserCenter
// @Router /user/update-position [put]
func (handler *Handler) UpdatePositions(c *gin.Context) {
	var req UpdatePositionsRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdatePositions(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdateRoles 替换请求指定用户的角色关联。
//
// 成功后 Service 撤销旧会话权限快照。
// @Tags UserCenter
// @Router /user/update-role [put]
func (handler *Handler) UpdateRoles(c *gin.Context) {
	var req UpdateRolesRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdateRoles(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdatePreferences 修改当前用户偏好配置。
//
// 配置字符串由 Service 原样保存。
// @Tags UserCenter
// @Router /user/update-preferences-config [put]
func (handler *Handler) UpdatePreferences(c *gin.Context) {
	var req UpdatePreferencesRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.UpdatePreferences(c.Request.Context(), utils.GetId(c), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// UpdateAvatar 上传并修改当前用户头像。
//
// 成功时沿用创建类响应的 HTTP 201 状态。
// @Tags UserCenter
// @Router /user/update-avatar [post]
func (handler *Handler) UpdateAvatar(c *gin.Context) {
	_, file, err := c.Request.FormFile("file")
	if err != nil {
		response.Error("接收文件失败", nil, c)
		return
	}
	if _, err = handler.service.UpdateAvatar(c.Request.Context(), utils.GetId(c), file); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("头像上传成功", c)
}

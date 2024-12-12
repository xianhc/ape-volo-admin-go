package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/auth"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/redis"
	"go.uber.org/zap"
)

type UserApi struct{}

// Create
// @Tags   User
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateUserReq true "CreateUpdateUserReq object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /user/create [post]
func (u *UserApi) Create(c *gin.Context) {
	var req dto.CreateUpdateUserReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(req)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	req.SetCreateBy(utils.GetAccount(c))
	err = userService.Create(&req)
	if err != nil {
		global.Logger.Error("创建失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   User
// @Summary 修改
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateUserReq true "CreateUpdateUserReq object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /user/edit [put]
func (u *UserApi) Update(c *gin.Context) {
	var req dto.CreateUpdateUserReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(req)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	req.SetUpdateBy(utils.GetAccount(c))
	err = userService.Update(&req)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   User
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /user/delete [delete]
func (u *UserApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = userService.Delete(idArray, utils.GetId(c), utils.GetAccount(c))
	if err != nil {
		global.Logger.Error("删除失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags      UserApe
// @Summary   分页获取用户列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.Pagination    true  "页码, 每页大小"
// @Success   200   {object}  response.Response{data=response.ActionResultPage,msg=string}  "分页获取用户列表,返回包括列表,总数,页码,每页数量"
// @Router    /user/query [get]
func (u *UserApi) Query(c *gin.Context) {
	var reqInfo dto.UserQueryCriteria
	err := c.ShouldBindQuery(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(reqInfo.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	var loginUserInfo *auth.LoginUserInfo
	_ = redis.Get(cachePrefix.OnlineKey+utils.MD5(ext.StringReplace(utils.GetToken(c), "Bearer ", "", -1)), &loginUserInfo)

	list := make([]permission.User, 0)
	var count int64
	err = userService.Query(&reqInfo, utils.GetId(c), loginUserInfo.DeptId, &list, &count)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}

	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

func (u *UserApi) Download(c *gin.Context) {
	var reqInfo dto.UserQueryCriteria
	err := c.ShouldBindQuery(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	var loginUserInfo auth.LoginUserInfo
	_ = redis.Get(cachePrefix.OnlineKey+utils.MD5(utils.GetToken(c)), &loginUserInfo)
	filePath, fileName, err := userService.Download(&reqInfo, utils.GetId(c), loginUserInfo.DeptId)
	if err != nil {
		global.Logger.Error("导出失败!", zap.Error(err))
		response.Error("导出失败", nil, c)
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

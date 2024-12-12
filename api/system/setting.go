package system

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type SettingApi struct{}

// Create
// @Tags   Setting
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateSettingDto true "CreateUpdateSettingDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /setting/create [post]
func (s *SettingApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateSettingDto
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
	err = settingService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Setting
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateSettingDto true "CreateUpdateSettingDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /setting/edit [put]
func (s *SettingApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateSettingDto
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
	err = settingService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Setting
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /setting/edit [delete]
func (s *SettingApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = settingService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Setting
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.SettingQueryCriteria true "Setting request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /setting/query [get]
func (s *SettingApi) Query(c *gin.Context) {
	var pageInfo dto.SettingQueryCriteria
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
	list := make([]system.Setting, 0)
	var count int64
	err = settingService.Query(&pageInfo, &list, &count)
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
// @Tags   Setting
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.SettingQueryCriteria true "Setting request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /setting/download [get]
func (s *SettingApi) Download(c *gin.Context) {
	var pageInfo dto.SettingQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	filePath, fileName, err := settingService.Download(&pageInfo)
	if err != nil {
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

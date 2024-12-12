package system

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
)

type FileRecordApi struct{}

// Upload
// @Tags      Storage
// @Summary   上传文件
// @Security  ApiKeyAuth
// @accept    multipart/form-data
// @Produce   application/json
// @Param     file  formData  file   true  "上传文件示例"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /storage/upload [post]
func (f *FileRecordApi) Upload(c *gin.Context) {
	_, file, err := c.Request.FormFile("file")
	if err != nil {
		response.Error("接收文件失败", nil, c)
		return
	}
	description := c.Query("description")
	if description == "" {
		response.Error("文件描述不能为空", nil, c)
		return
	}
	var fileSize = global.Config.System.FileLimitSize * 1024 * 1024
	if file.Size > fileSize {
		response.Error(fmt.Sprintf("文件过大，请选择文件小于等于%dMB的重新进行尝试!", fileSize), nil, c)
		return
	}
	createUpdateFileRecordDto := &dto.CreateUpdateFileRecordDto{Description: description}
	createUpdateFileRecordDto.SetCreateBy(utils.GetAccount(c))
	err = fileRecordService.Create(*createUpdateFileRecordDto, file) // 文件上传后拿到文件路径
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Storage
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateFileRecordDto true "CreateUpdateFileRecordDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /storage/edit [put]
func (f *FileRecordApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateFileRecordDto
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
	err = fileRecordService.Update(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Storage
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /storage/delete [delete]
func (f *FileRecordApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = fileRecordService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   FileRecord
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.FileRecordQueryCriteria true "FileRecord request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /storage/query [get]
func (f *FileRecordApi) Query(c *gin.Context) {
	var pageInfo dto.FileRecordQueryCriteria
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
	list := make([]system.FileRecord, 0)
	var count int64
	err = fileRecordService.Query(&pageInfo, &list, &count)
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
// @Tags   FileRecord
// @Summary 导出
// @Accept json
// @Produce json
// @Param request body dto.FileRecordQueryCriteria true "FileRecord request object"
// @Success 200 {object} response.ActionResultPage "导出成功"
// @Router /storage/download [get]
func (f *FileRecordApi) Download(c *gin.Context) {
	var pageInfo dto.FileRecordQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	filePath, fileName, err := fileRecordService.Download(&pageInfo)
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

package filerecord

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理文件记录的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入统一响应。
type Handler struct {
	service Service
}

// Upload 接收文件并创建文件记录。
//
// 方法校验描述和文件大小后把 multipart 文件交给 Service 保存。
// @Tags Storage
// @Summary 上传文件
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param file formData file true "上传文件示例"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /file-record/create [post]
func (handler *Handler) Upload(c *gin.Context) {
	_, file, err := c.Request.FormFile("file")
	if err != nil {
		response.Error("接收文件失败", nil, c)
		return
	}
	description := c.Query("description")
	fileSize := global.Config.System.FileLimitSize * 1024 * 1024
	if file.Size > fileSize {
		response.Error(fmt.Sprintf("文件过大，请选择文件小于等于%dMB的重新进行尝试!", global.Config.System.FileLimitSize), nil, c)
		return
	}
	req := CreateAndUpdateFileRecordRequest{Description: description}
	if err := handler.service.Create(c.Request.Context(), req, file); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改文件记录描述。
//
// 请求校验通过后交给 Service 处理，成功时返回 HTTP 204。
// @Tags Storage
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body filerecord.CreateAndUpdateFileRecordRequest true "CreateAndUpdateFileRecordRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /file-record/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateFileRecordRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组文件记录及其本地文件。
//
// 主键集合兼容对象和裸数组两种已有请求格式。
// @Tags Storage
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /file-record/delete [delete]
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

// Query 查询文件记录分页列表。
//
// 查询参数同时兼容 URL query 与请求体，成功时返回统一分页结构。
// @Tags FileRecord
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body filerecord.QueryFileRecordRequest true "FileRecord request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /file-record/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryFileRecordRequest
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

// Download 导出文件记录列表。
//
// 导出成功后以 XLSX 附件响应，筛选条件与列表查询保持一致。
// @Tags FileRecord
// @Summary 导出
// @Accept json
// @Produce json
// @Param request body filerecord.QueryFileRecordRequest true "FileRecord request object"
// @Success 200 {object} response.ActionResultPage "导出成功"
// @Router /file-record/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryFileRecordRequest
	if err := request.BindParams(c, &req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := handler.service.Download(c.Request.Context(), req)
	if err != nil {
		response.Error("导出失败", nil, c)
		return
	}
	response.DownloadFile(c, filePath, fileName)
}

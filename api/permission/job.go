package permission

import (
	"github.com/gin-gonic/gin"
	"go-apevolo/global"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go.uber.org/zap"
)

type JobApi struct{}

// Create
// @Tags   Job
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateJobDto true "CreateUpdateJobDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /job/create [post]
func (j *JobApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateJobDto
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
	err = jobService.Create(&reqInfo)
	if err != nil {
		global.Logger.Error("创建失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update
// @Tags   Job
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateJobDto true "CreateUpdateJobDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /job/edit [put]
func (j *JobApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateJobDto
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
	err = jobService.Update(&reqInfo)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete
// @Tags   Job
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /job/delete [delete]
func (j *JobApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	err = jobService.Delete(idArray, utils.GetAccount(c))
	if err != nil {
		global.Logger.Error("删除失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	response.Success("", c)
}

// Query
// @Tags   Job
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.JobQueryCriteria true "Job request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /job/query [get]
func (j *JobApi) Query(c *gin.Context) {
	var jobQueryCriteria dto.JobQueryCriteria
	err := c.ShouldBindQuery(&jobQueryCriteria)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	err = utils.VerifyData(jobQueryCriteria.Pagination)
	if err != nil {
		response.Error("", utils.GetVerifyErr(err), c)
		return
	}
	list := make([]permission.Job, 0)
	var count int64
	err = jobService.Query(&jobQueryCriteria, &list, &count)
	if err != nil {
		global.Logger.Error("获取失败!", zap.Error(err))
		response.Error("获取失败", nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

func (j *JobApi) Download(c *gin.Context) {
	var jobQueryCriteria dto.JobQueryCriteria
	err := c.ShouldBindQuery(&jobQueryCriteria)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	filePath, fileName, err := jobService.Download(&jobQueryCriteria)
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

// All
// @Tags   Job
// @Summary 查询全部岗位
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /job/queryAll [get]
func (j *JobApi) All(c *gin.Context) {
	jobs := make([]permission.Job, 0)
	err := jobService.GetAllJob(&jobs)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	total := len(jobs)
	response.ResultPage(response.ActionResultPage{
		Content:       jobs,
		TotalElements: int64(total),
	}, c)
}

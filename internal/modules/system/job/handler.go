package job

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
)

// Handler 处理调度任务配置与运行时操作的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Create 创建调度任务。
//
// Service 负责触发器组合校验、持久化和运行时注册。
// @Tags Job
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body job.CreateAndUpdateJobRequest true "CreateAndUpdateJobRequest object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /job-manager/create [post]
func (handler *Handler) Create(c *gin.Context) {
	var req CreateAndUpdateJobRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Create(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.Create("", c)
}

// Update 修改调度任务。
//
// Service 在写库后同步重建或移除运行时任务。
// @Tags Job
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body job.CreateAndUpdateJobRequest true "CreateAndUpdateJobRequest object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /job-manager/update [put]
func (handler *Handler) Update(c *gin.Context) {
	var req CreateAndUpdateJobRequest
	if !request.BindAndValidateJSON(c, &req) {
		return
	}
	if err := handler.service.Update(c.Request.Context(), req); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Delete 删除一组调度任务。
//
// Service 同步删除数据库配置与运行时任务。
// @Tags Job
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /job-manager/delete [delete]
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

// Query 查询调度任务分页列表。
//
// 每条结果包含当前运行时触发器状态。
// @Tags Job
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body job.QueryJobRequest true "QueryJobRequest object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /job/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryJobRequest
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
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{Items: items, TotalCount: total}, c)
}

// Download 导出调度任务列表。
//
// 导出成功后以 XLSX 附件响应。
// @Tags Job
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body job.QueryJobRequest true "Task request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /job/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryJobRequest
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

// Execute 注册并启动指定调度任务。
//
// 主键继续从 query 参数读取，成功时返回 HTTP 204。
// @Tags Job
// @Summary 执行
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "执行成功"
// @Router /job-manager/execute [post]
func (handler *Handler) Execute(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.Execute(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Pause 暂停指定调度任务。
//
// 只有运行中任务可以暂停。
// @Tags Job
// @Summary 暂停
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "暂停成功"
// @Router /job-manager/pause [post]
func (handler *Handler) Pause(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.Pause(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Resume 恢复指定调度任务。
//
// 只有暂停状态的任务可以恢复。
// @Tags Job
// @Summary 暂停
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "暂停成功"
// @Router /job-manager/resume [post]
func (handler *Handler) Resume(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.Resume(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Stop 停用配置并停止指定作业的后续调度。
//
// 已经开始的执行不会被中断，成功时返回 HTTP 204。
// @Tags Job
// @Summary 停止调度
// @Accept json
// @Produce json
// @Success 204 {object} response.ActionResult "停止成功"
// @Router /job-manager/stop [post]
func (handler *Handler) Stop(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.Stop(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// Restart 按数据库中的最新配置重新调度指定作业。
//
// 该操作同时启用配置，成功时返回 HTTP 204。
// @Tags Job
// @Summary 重新调度
// @Accept json
// @Produce json
// @Success 204 {object} response.ActionResult "重新调度成功"
// @Router /job-manager/restart [post]
func (handler *Handler) Restart(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	if err := handler.service.Restart(c.Request.Context(), ext.StringToInt64(id)); err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	response.NoContent(c)
}

// All 查询全部调度任务。
//
// 该接口直接返回 Response 数组，供前端下拉或管理视图使用。
// @Tags Job
// @Summary 查询全部定时任务
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} []job.JobResponse "任务列表"
// @Router /job/get-all [get]
func (handler *Handler) All(c *gin.Context) {
	items, err := handler.service.All(c.Request.Context())
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	c.JSON(http.StatusOK, items)
}

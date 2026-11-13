package scheduledjoblog

import (
	"github.com/gin-gonic/gin"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/request"
	"github.com/xianhc/ape-volo-admin-go/internal/payloads/response"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// Handler 处理调度任务日志的 HTTP 请求。
//
// Handler 只负责参数绑定、校验、调用 Service 和写入响应。
type Handler struct{ service Service }

// Query 查询调度任务日志分页列表。
//
// 查询支持任务名称、任务组、执行结果和排序条件。
// @Tags ScheduledJobLog
// @Summary 分页获取任务执行日志
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query scheduledjoblog.QueryScheduledJobLogRequest true "任务名称, 任务分组, 是否成功, 页码, 每页大小"
// @Success 200 {object} response.Response{data=response.ActionResultPage} "任务执行日志列表"
// @Router /job-log/get-list [get]
func (handler *Handler) Query(c *gin.Context) {
	var req QueryScheduledJobLogRequest
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

// Download 导出调度任务日志。
//
// 导出成功后以 XLSX 附件响应。
// @Tags ScheduledJobLog
// @Summary 导出任务执行日志
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query scheduledjoblog.QueryScheduledJobLogRequest false "查询条件"
// @Success 200 {string} string "文件流"
// @Router /job-log/download [post]
func (handler *Handler) Download(c *gin.Context) {
	var req QueryScheduledJobLogRequest
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

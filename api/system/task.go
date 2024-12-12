package system

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go-apevolo/global"
	"go-apevolo/job"
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/payloads/response"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/timer"
	"go.uber.org/zap"
)

type TaskApi struct{}

// Create
// @Tags   Task
// @Summary 创建
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateTaskDto true "CreateUpdateTaskDto object"
// @Success 201 {object} response.ActionResult "创建成功"
// @Router /tasks/create [post]
func (t *TaskApi) Create(c *gin.Context) {
	var reqInfo dto.CreateUpdateTaskDto
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
	if reqInfo.TriggerType == timer.Cron {
		if reqInfo.Cron == "" {
			response.Error("cron模式下请设置作业执行cron表达式", nil, c)
			return
		} else {
			_, err = cron.ParseStandard(reqInfo.Cron)
			if err != nil {
				response.Error(err.Error(), nil, c)
				return
			}
		}
	} else if reqInfo.TriggerType == timer.Simple {
		if *reqInfo.IntervalSecond <= 5 {
			response.Error("simple模式下请设置作业间隔执行秒数", nil, c)
			return
		}
	}

	reqInfo.SetCreateBy(utils.GetAccount(c))
	err = taskService.Create(&reqInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	if reqInfo.IsEnable {
		// 通过反射获取函数名对应的函数值
		fn := job.GlobalFuncs[reqInfo.AssemblyName+reqInfo.ClassName]
		if fn != nil {
			taskJob := timer.TaskJob{
				Id:             0,
				TriggerType:    reqInfo.TriggerType,
				TaskName:       reqInfo.TaskName,
				TaskGroup:      reqInfo.TaskGroup,
				Cron:           reqInfo.Cron,
				ClassName:      reqInfo.ClassName,
				AssemblyName:   reqInfo.AssemblyName,
				StartTime:      reqInfo.StartTime,
				EndTime:        reqInfo.EndTime,
				IntervalSecond: reqInfo.IntervalSecond,
				IsEnable:       reqInfo.IsEnable,
			}
			_, err = global.Timer.AddTaskByFunc(taskJob, fn)
			if err != nil {
				global.Logger.Error(err.Error(), zap.Error(err))
			}
		}
	}

	response.Create("", c)
}

// Update
// @Tags   Task
// @Summary 编辑
// @Accept json
// @Produce json
// @Param request body dto.CreateUpdateTaskDto true "CreateUpdateTaskDto object"
// @Success 204 {object} response.ActionResult "编辑成功"
// @Router /tasks/edit [put]
func (t *TaskApi) Update(c *gin.Context) {
	var reqInfo dto.CreateUpdateTaskDto
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
	if reqInfo.TriggerType == timer.Cron {
		if reqInfo.Cron == "" {
			response.Error("cron模式下请设置作业执行cron表达式", nil, c)
			return
		} else {
			_, err = cron.ParseStandard(reqInfo.Cron)
			if err != nil {
				response.Error(err.Error(), nil, c)
				return
			}
		}
	} else if reqInfo.TriggerType == timer.Simple {
		if *reqInfo.IntervalSecond <= 10 {
			response.Error("simple模式下请设置作业间隔执行秒数", nil, c)
			return
		}
	}

	reqInfo.SetUpdateBy(utils.GetAccount(c))
	err = taskService.Update(&reqInfo)
	if err != nil {
		global.Logger.Error("修改失败!", zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	taskJob := timer.TaskJob{
		Id:             0,
		TriggerType:    reqInfo.TriggerType,
		TaskName:       reqInfo.TaskName,
		TaskGroup:      reqInfo.TaskGroup,
		Cron:           reqInfo.Cron,
		ClassName:      reqInfo.ClassName,
		AssemblyName:   reqInfo.AssemblyName,
		StartTime:      reqInfo.StartTime,
		EndTime:        reqInfo.EndTime,
		IntervalSecond: reqInfo.IntervalSecond,
		IsEnable:       reqInfo.IsEnable,
	}
	if reqInfo.IsEnable {
		global.Timer.Remove(taskJob.TaskName)
		global.Timer.Delete(taskJob.TaskName)
		// 通过反射获取函数名对应的函数值
		fn := job.GlobalFuncs[reqInfo.AssemblyName+reqInfo.ClassName]
		if fn != nil {
			_, err = global.Timer.AddTaskByFunc(taskJob, fn)
			if err != nil {
				global.Logger.Error(err.Error(), zap.Error(err))
			}
		}
	} else {
		global.Timer.Remove(taskJob.TaskName)
		global.Timer.Delete(taskJob.TaskName)
	}
	response.NoContent(c)
}

// Delete
// @Tags   Task
// @Summary 删除
// @Accept json
// @Produce json
// @Param request body request.IdCollection true "ID数组"
// @Success 200 {object} response.ActionResult "删除成功"
// @Router /tasks/delete [delete]
func (t *TaskApi) Delete(c *gin.Context) {
	var idArray request.IdCollection
	err := c.ShouldBindJSON(&idArray)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	list := make([]system.Task, 0)
	err = taskService.Delete(idArray, utils.GetAccount(c), &list)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}
	for _, item := range list {
		global.Timer.Remove(item.TaskName)
		global.Timer.Delete(item.TaskName)
	}

	response.Success("", c)
}

// Query
// @Tags   Task
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.TaskQueryCriteria true "TaskQueryCriteria object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /tasks/query [get]
func (t *TaskApi) Query(c *gin.Context) {
	var pageInfo dto.TaskQueryCriteria
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
	list := make([]system.Task, 0)
	var count int64
	err = taskService.Query(&pageInfo, &list, &count)
	if err != nil {
		response.Error("获取失败:"+err.Error(), nil, c)
		return
	}
	response.ResultPage(response.ActionResultPage{
		Content:       list,
		TotalElements: count,
	}, c)
}

// Download
// @Tags   Task
// @Summary 查询
// @Accept json
// @Produce json
// @Param request body dto.TaskQueryCriteria true "Task request object"
// @Success 200 {object} response.ActionResultPage "查询成功"
// @Router /task/download [get]
func (t *TaskApi) Download(c *gin.Context) {
	var pageInfo dto.TaskQueryCriteria
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.Error(err.Error(), nil, c)
		return
	}

	filePath, fileName, err := taskService.Download(&pageInfo)
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

// Execute
// @Tags   Task
// @Summary 执行
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "执行成功"
// @Router /tasks/execute [put]
func (t *TaskApi) Execute(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	var newId = ext.StringToInt64(id)
	task := system.Task{}
	err := taskService.QueryFirst(newId, &task)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}

	_, b := global.Timer.FindTaskStatus(task.TaskName)
	if !b {
		// 通过反射获取函数名对应的函数值
		fn := job.GlobalFuncs[task.AssemblyName+task.ClassName]
		if fn != nil {
			if !task.IsEnable {
				createUpdateTaskDto := dto.CreateUpdateTaskDto{
					IsEnable: true,
					RootKey:  model.RootKey{Id: newId},
				}
				createUpdateTaskDto.SetUpdateBy(utils.GetAccount(c))
				err = taskService.Update(&createUpdateTaskDto)
				if err != nil {
					global.Logger.Error(err.Error(), zap.Error(err))
					response.Error(err.Error(), nil, c)
					return
				}
			}
			taskJob := timer.TaskJob{
				Id:             0,
				TriggerType:    task.TriggerType,
				TaskName:       task.TaskName,
				TaskGroup:      task.TaskGroup,
				Cron:           task.Cron,
				ClassName:      task.ClassName,
				AssemblyName:   task.AssemblyName,
				StartTime:      task.StartTime,
				EndTime:        task.EndTime,
				IntervalSecond: task.IntervalSecond,
				IsEnable:       task.IsEnable,
			}
			_, err = global.Timer.AddTaskByFunc(taskJob, fn)
			if err != nil {
				global.Logger.Error(err.Error(), zap.Error(err))
				response.Error(err.Error(), nil, c)
				return
			} else {
				response.NoContent(c)
				return
			}
		} else {
			response.Error("作业函数不存在,请检查", nil, c)
			return
		}
	}

	response.Error("执行失败", nil, c)
}

// Pause
// @Tags   Task
// @Summary 暂停
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "暂停成功"
// @Router /tasks/pause [put]
func (t *TaskApi) Pause(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	task := system.Task{}
	err := taskService.QueryFirst(ext.StringToInt64(id), &task)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	taskStatus, b := global.Timer.FindTaskStatus(task.TaskName)
	if b {
		if taskStatus.State == "运行中" {
			global.Timer.StopTask(task.TaskName)
			response.NoContent(c)
			return
		}
	}
	response.Error("暂停失败", nil, c)
}

// Resume
// @Tags   Task
// @Summary 暂停
// @Accept json
// @Produce json
// @Success 200 {object} response.ActionResult "暂停成功"
// @Router /tasks/resume [put]
func (t *TaskApi) Resume(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error("id is null", nil, c)
		return
	}
	task := system.Task{}
	err := taskService.QueryFirst(ext.StringToInt64(id), &task)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		response.Error(err.Error(), nil, c)
		return
	}
	taskStatus, b := global.Timer.FindTaskStatus(task.TaskName)
	if b {
		if taskStatus.State == "暂停" {
			global.Timer.StartTask(task.TaskName)
			response.NoContent(c)
			return
		}
	}
	response.Error("恢复失败", nil, c)
}

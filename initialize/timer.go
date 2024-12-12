package initialize

import (
	"fmt"
	"go-apevolo/global"
	"go-apevolo/job"
	"go-apevolo/model/system"
	"go-apevolo/service"
	"go-apevolo/utils/timer"
	"go.uber.org/zap"
)

var taskService = service.ServiceGroupApp.SystemServiceGroup.TaskService

func Timer() {
	job.RegisterTaskFuncs()
	taskList := make([]system.Task, 0)
	err := taskService.QueryAll(&taskList)
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
		return
	}

	// 定义一个目标结构体的切片
	var taskJobs []timer.TaskJob

	// 遍历源结构体的切片，将每个元素的字段值复制到目标结构体的切片中
	for _, item := range taskList {
		// 创建一个新的目标结构体，并赋值
		taskJob := timer.TaskJob{
			Id:             item.Id,
			TriggerType:    item.TriggerType,
			TaskName:       item.TaskName,
			Cron:           item.Cron,
			ClassName:      item.ClassName,
			AssemblyName:   item.AssemblyName,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			IntervalSecond: item.IntervalSecond,
			IsEnable:       item.IsEnable,
		}
		// 将目标结构体添加到目标切片中
		taskJobs = append(taskJobs, taskJob)
	}

	for _, task := range taskJobs {
		if !task.IsEnable {
			continue
		}
		// 通过反射获取函数名对应的函数值
		fn := job.GlobalFuncs[task.AssemblyName+task.ClassName]
		if fn == nil {
			global.Logger.Error(fmt.Sprintf("任务 %s 不存在\n", task.TaskName))
			continue
		}

		_, err := global.Timer.AddTaskByFunc(task, fn)
		if err != nil {
			global.Logger.Error(err.Error(), zap.Error(err))
		}

	}
}

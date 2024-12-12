package system

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type TaskService struct{}

// Create
// @description: 创建
// @receiver: taskService
// @param: req
// @return: error
func (taskService *TaskService) Create(req *dto.CreateUpdateTaskDto) error {
	task := &system.Task{}
	req.Generate(task)
	err := global.Db.Create(task).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: taskService
// @param: req
// @return: error
func (taskService *TaskService) Update(req *dto.CreateUpdateTaskDto) error {
	oldTask := &system.Task{}
	task := &system.Task{}
	err := global.Db.Scopes(utils.IsDeleteSoft).Where("id = ?", req.Id).First(oldTask).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	req.Generate(task)
	err = global.Db.Updates(task).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: taskService
// @param: idArray
// @param: updateBy
// @param: tasks
// @return: error
func (taskService *TaskService) Delete(idArray request.IdCollection, updateBy string, tasks *[]system.Task) error {
	//var tasks []system.Task
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&tasks, "id in (?)", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(*tasks) == 0 {
		return errors.New("数据不存在")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&system.Task{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		system.Task{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// QueryFirst
// @description: 查询
// @receiver: taskService
// @param: id
// @param: task
// @return: error
func (taskService *TaskService) QueryFirst(id int64, task *system.Task) error {
	err := global.Db.Scopes(utils.IsDeleteSoft).First(task, id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: taskService
// @param: info
// @param: list
// @param: count
// @return: error
func (taskService *TaskService) Query(info *dto.TaskQueryCriteria, list *[]system.Task, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildTaskQuery(global.Db.Model(&system.Task{}), info)
	err := db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}

	for _, task := range *list {
		task.TriggerStatus = global.Timer.GetTaskStatus(task.TaskName)
	}
	return nil
}

// QueryAll
// @description: 查询全部
// @receiver: taskService
// @param: list
// @return: error
func (taskService *TaskService) QueryAll(list *[]system.Task) error {
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: taskService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (taskService *TaskService) Download(info *dto.TaskQueryCriteria) (filePath string, fileName string, err error) {
	var tasks []system.Task
	// 创建db并构建查询条件
	err = buildTaskQuery(global.Db.Model(&system.Task{}), info).Find(&tasks).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Tasks")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "任务名称", "任务组", "Cron表达式", "程序集名称", "执行类", "描述", "负责人", "告警邮箱", "失败是否继续", "执行次数", "触发器模式", "执行间隔时间", "循环执行次数", "是否启动",
		"执行传参", "触发器状态", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, task := range tasks {
		row := sheet.AddRow()
		intervalSecondStr := "0"
		cycleRunTimesStr := "0"
		if task.IntervalSecond != nil {
			intervalSecondStr = strconv.Itoa(int(*task.IntervalSecond))
		}
		if task.CycleRunTimes != nil {
			cycleRunTimesStr = strconv.Itoa(int(*task.CycleRunTimes))
		}

		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(task.Id, 10),
			task.TaskName,
			task.TaskGroup,
			task.Cron,
			task.AssemblyName,
			task.ClassName,
			task.Description,
			task.Principal,
			task.AlertEmail,
			func() string {
				if task.PauseAfterFailure {
					return "是"
				}
				return "否"
			}(),
			task.RunTimes,
			//*task.StartTime,
			//*task.EndTime,
			func() string {
				if task.TriggerType == 1 {
					return "cron"
				}
				return "simple"
			}(),
			intervalSecondStr,
			cycleRunTimesStr,
			func() string {
				if task.IsEnable {
					return "启用"
				}
				return "禁用"
			}(),
			task.RunParams,
			task.TriggerStatus,
			task.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Task_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// buildTaskQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildTaskQuery(db *gorm.DB, info *dto.TaskQueryCriteria) *gorm.DB {
	if info.TaskName != "" {
		db = db.Where("task_name LIKE ? ", "%"+info.TaskName+"%")
	}
	//if len(info.IdArray) > 0 {
	//	db = db.Where("id in (?) ", info.IdArray)
	//}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}

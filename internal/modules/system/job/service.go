package job

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	jobmodule "github.com/xianhc/ape-volo-admin-go/internal/job"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	timerutil "github.com/xianhc/ape-volo-admin-go/internal/utils/timer"
	"go.uber.org/zap"
)

// Service 实现调度任务配置与运行时编排业务。
//
// Service 通过 Repository 持久化配置，并同步维护进程内调度器。
type Service struct{ repository repository }

var runtimeMu sync.Mutex

// Create 创建调度任务并按需注册运行时任务。
//
// 触发器配置验证通过后先写库，启用且函数存在时再注册调度器。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateJobRequest) error {
	if req.Id == 0 {
		req.Id = int64(utils.GenerateID())
	}
	if err := utils.InTransaction(ctx, func(ctx context.Context) error { return service.create(ctx, req) }); err != nil {
		return err
	}
	if req.Enabled {
		return service.applyRuntime(ctx, req.Id)
	}
	return nil
}

func (service *Service) create(ctx context.Context, req CreateAndUpdateJobRequest) error {
	task := entityFromRequest(req)
	if task.TriggerType == timerutil.Simple && (task.IntervalSecond == nil || *task.IntervalSecond <= 5) {
		return errors.New("simple模式下请设置作业间隔执行秒数！")
	}
	if err := service.validateUnique(ctx, *task); err != nil {
		return err
	}
	if req.Enabled {
		if err := validateDefinition(definitionFromEntity(*task)); err != nil {
			return err
		}
	}
	if err := service.repository.Create(ctx, task); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改调度任务并同步运行时任务。
//
// 更新成功后删除旧调度实例，启用时按新配置重新注册。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateJobRequest) error {
	if err := utils.InTransaction(ctx, func(ctx context.Context) error { return service.update(ctx, req) }); err != nil {
		return err
	}
	return service.applyRuntime(ctx, req.Id)
}

func (service *Service) update(ctx context.Context, req CreateAndUpdateJobRequest) error {
	_, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	task := entityFromRequest(req)
	if err := service.validateUnique(ctx, *task); err != nil {
		return err
	}
	if req.Enabled {
		if err := validateDefinition(definitionFromEntity(*task)); err != nil {
			return err
		}
	}
	if err := service.repository.Update(ctx, task); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 删除调度任务并移除运行时实例。
//
// 数据库中没有匹配任务时返回“数据不存在”。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.delete(ctx, ids) })
}

func (service *Service) delete(ctx context.Context, ids []string) error {
	tasks, err := service.repository.FindByIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if len(tasks) == 0 {
		return errors.New("数据不存在")
	}
	for _, task := range tasks {
		if !global.Timer.DeleteByID(task.Id) {
			return errors.New("停止调度失败，作业配置已保留，请重试删除。")
		}
	}
	if _, err := service.repository.Delete(ctx, ids); err != nil {
		return errors.New("调度已停止，但删除配置失败，请重试删除。")
	}
	return nil
}

// Query 查询调度任务分页列表。
//
// 每条 Response 注入当前运行时触发器状态。
func (service *Service) Query(ctx context.Context, req QueryJobRequest) ([]JobResponse, int64, error) {
	tasks, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	result := make([]JobResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, fromEntity(task, int32(global.Timer.GetTaskStatus(task.Id))))
	}
	return result, total, nil
}

// All 查询全部调度任务用于 HTTP 返回。
//
// 返回独立 Response 数组，不暴露数据库 Entity。
func (service *Service) All(ctx context.Context) ([]JobResponse, error) {
	tasks, err := service.repository.FindAll(ctx, queryOptions{})
	if err != nil {
		return nil, logDatabaseError(err)
	}
	result := make([]JobResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, fromEntity(task, int32(global.Timer.GetTaskStatus(task.Id))))
	}
	return result, nil
}

// Definitions 查询全部调度任务的内部运行时快照。
//
// 启动初始化使用该 DTO 注册任务，不接触数据库 Entity。
func (service *Service) Definitions(ctx context.Context) ([]jobmodule.Definition, error) {
	tasks, err := service.repository.FindAll(ctx, queryOptions{})
	if err != nil {
		return nil, logDatabaseError(err)
	}
	result := make([]jobmodule.Definition, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, definitionFromEntity(task))
	}
	return result, nil
}

// Download 导出符合条件的调度任务。
//
// 导出保持现有字段顺序和值格式。
func (service *Service) Download(ctx context.Context, req QueryJobRequest) (filePath, fileName string, err error) {
	tasks, err := service.repository.FindAll(ctx, listOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	file := excel.NewFile("Tasks")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"ID", "任务名称", "任务组", "Cron表达式", "程序集名称", "执行类", "描述", "负责人", "告警邮箱", "失败是否继续", "触发器模式", "执行间隔时间", "循环执行次数", "是否启动", "执行传参", "触发器状态", "创建时间"})
	for _, task := range tasks {
		interval, cycle := "0", "0"
		if task.IntervalSecond != nil {
			interval = strconv.Itoa(int(*task.IntervalSecond))
		}
		if task.CycleRunTimes != nil {
			cycle = strconv.Itoa(int(*task.CycleRunTimes))
		}
		pause, trigger, enabled := "否", "simple", "禁用"
		if task.PauseAfterFailure {
			pause = "是"
		}
		if task.TriggerType == 1 {
			trigger = "cron"
		}
		if task.Enabled {
			enabled = "启用"
		}
		file.AddRow(strconv.FormatInt(task.Id, 10), task.JobName, task.JobGroup, task.Cron, task.AssemblyName, task.ClassName, task.Description, task.Principal, task.AlertEmail, pause, trigger, interval, cycle, enabled, task.RunParams, int32(global.Timer.GetTaskStatus(task.Id)), task.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "Task_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}

// Execute 启动尚未注册的调度任务。
//
// 未启用任务会先持久化启用状态，再注册运行时执行体。
func (service *Service) Execute(ctx context.Context, id int64) error {
	if err := utils.InTransaction(ctx, func(ctx context.Context) error { return service.execute(ctx, id) }); err != nil {
		return err
	}
	return service.applyRuntime(ctx, id)
}

func (service *Service) execute(ctx context.Context, id int64) error {
	task, err := service.repository.FindByID(ctx, id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if global.Timer.GetTaskStatus(task.Id) != timerutil.TriggerStateNone {
		return errors.New("已有调度计划，请使用恢复或重新调度。")
	}
	definition := definitionFromEntity(task)
	if err := validateDefinition(definition); err != nil {
		return err
	}
	if !task.Enabled {
		if err := service.repository.SetEnabled(ctx, task.Id, true); err != nil {
			return logDatabaseError(err)
		}
		definition.Enabled = true
	}
	return nil
}

// Pause 暂停正在运行的调度任务。
//
// 任务不存在或状态不是“运行中”时返回“暂停失败”。
func (service *Service) Pause(ctx context.Context, id int64) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.pause(ctx, id) })
}

func (service *Service) pause(ctx context.Context, id int64) error {
	task, err := service.repository.FindByID(ctx, id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if global.Timer.GetTaskStatus(task.Id) == timerutil.TriggerStateNormal && global.Timer.StopTask(task.Id) {
		return nil
	}
	return errors.New("暂停失败，请重试！")
}

// Resume 恢复已暂停的调度任务。
//
// 任务不存在或状态不是“暂停”时返回“恢复失败”。
func (service *Service) Resume(ctx context.Context, id int64) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.resume(ctx, id) })
}

func (service *Service) resume(ctx context.Context, id int64) error {
	task, err := service.repository.FindByID(ctx, id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if global.Timer.GetTaskStatus(task.Id) == timerutil.TriggerStatePaused && global.Timer.StartTask(task.Id) {
		return nil
	}
	return errors.New("恢复失败，请重试！")
}

// Stop 停用配置并移除后续调度计划。
//
// 已经开始的执行由调度器按正常生命周期完成。
func (service *Service) Stop(ctx context.Context, id int64) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error {
		task, err := service.repository.FindByID(ctx, id)
		if errors.Is(err, errNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		if err != nil {
			return logDatabaseError(err)
		}
		if !global.Timer.DeleteByID(task.Id) {
			return errors.New("停止调度失败，作业配置已保留，请重试。")
		}
		if err := service.repository.SetEnabled(ctx, task.Id, false); err != nil {
			return errors.New("调度已停止，但停用配置失败，请重试。")
		}
		return nil
	})
}

// Restart 启用配置并按数据库中的最新内容重新调度。
//
// 同一作业 ID 的旧计划会在事务提交后被幂等替换。
func (service *Service) Restart(ctx context.Context, id int64) error {
	if err := utils.InTransaction(ctx, func(ctx context.Context) error {
		task, err := service.repository.FindByID(ctx, id)
		if errors.Is(err, errNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		if err != nil {
			return logDatabaseError(err)
		}
		definition := definitionFromEntity(task)
		definition.Enabled = true
		if err := validateDefinition(definition); err != nil {
			return err
		}
		if err := service.repository.SetEnabled(ctx, task.Id, true); err != nil {
			return logDatabaseError(err)
		}
		return nil
	}); err != nil {
		return err
	}
	return service.applyRuntime(ctx, id)
}

func validateDefinition(definition jobmodule.Definition) error {
	if err := timerutil.Validate(timerDefinition(definition)); err != nil {
		return err
	}
	_, err := jobmodule.Prepare(definition)
	return err
}

func (service *Service) deferRuntime(ctx context.Context, id int64) {
	utils.AfterCommit(ctx, "runtime:"+strconv.FormatInt(id, 10), func(ctx context.Context) {
		if err := service.syncRuntime(ctx, id); err != nil {
			global.Logger.Error("更新任务调度失败", zap.Int64("id", id), zap.Error(err))
		}
	})
}

func (service *Service) applyRuntime(ctx context.Context, id int64) error {
	if utils.TxFromContext(ctx) != nil && !utils.TxSkipped(ctx) {
		service.deferRuntime(ctx, id)
		return nil
	}
	if err := service.syncRuntime(ctx, id); err != nil {
		global.Logger.Error("更新任务调度失败", zap.Int64("id", id), zap.Error(err))
		return errors.New("作业配置已保存，但调度同步失败，请重新保存或执行。")
	}
	return nil
}

func (service *Service) syncRuntime(ctx context.Context, id int64) error {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	// 回调可能晚于后续事务完成，始终以最新已提交配置为准。
	task, err := service.repository.FindByID(utils.WithoutDataScope(ctx), id)
	if errors.Is(err, errNotFound) {
		if global.Timer.DeleteByID(id) {
			return nil
		}
		return errors.New("任务调度器已关闭")
	}
	if err != nil {
		return err
	}
	if !task.Enabled {
		if global.Timer.DeleteByID(id) {
			return nil
		}
		return errors.New("任务调度器已关闭")
	}
	return service.RegisterRuntime(definitionFromEntity(task))
}

// RegisterRuntime 注册已提交或启动时加载的任务配置。
func (service *Service) RegisterRuntime(def jobmodule.Definition) error {
	fn, err := jobmodule.Prepare(def)
	if err != nil {
		return err
	}
	_, err = global.Timer.AddTaskByFunc(timerDefinition(def), fn)
	return err
}
func timerDefinition(def jobmodule.Definition) timerutil.ScheduledJob {
	var startTime, endTime *time.Time
	if def.StartTime != nil {
		value := def.StartTime.Time
		startTime = &value
	}
	if def.EndTime != nil {
		value := def.EndTime.Time
		endTime = &value
	}
	return timerutil.ScheduledJob{Id: def.Id, TriggerType: def.TriggerType, JobName: def.JobName, JobGroup: def.JobGroup, Cron: def.Cron, ClassName: def.ClassName, AssemblyName: def.AssemblyName, IntervalSecond: def.IntervalSecond, CycleRunTimes: def.CycleRunTimes, StartTime: startTime, EndTime: endTime, PauseAfterFailure: def.PauseAfterFailure, Enabled: def.Enabled}
}
func definitionFromEntity(task Job) jobmodule.Definition {
	return jobmodule.Definition{Id: task.Id, JobName: task.JobName, JobGroup: task.JobGroup, Cron: task.Cron, ClassName: task.ClassName, AssemblyName: task.AssemblyName, Description: task.Description, Principal: task.Principal, AlertEmail: task.AlertEmail, PauseAfterFailure: task.PauseAfterFailure, StartTime: task.StartTime, EndTime: task.EndTime, TriggerType: task.TriggerType, IntervalSecond: task.IntervalSecond, CycleRunTimes: task.CycleRunTimes, Enabled: task.Enabled, RunParams: task.RunParams}
}
func entityFromRequest(req CreateAndUpdateJobRequest) *Job {
	task := &Job{
		JobName:           req.JobName,
		JobGroup:          req.JobGroup,
		Cron:              req.Cron,
		ClassName:         req.ClassName,
		AssemblyName:      req.AssemblyName,
		Description:       req.Description,
		Principal:         req.Principal,
		AlertEmail:        req.AlertEmail,
		PauseAfterFailure: req.PauseAfterFailure,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		TriggerType:       req.TriggerType,
		IntervalSecond:    req.IntervalSecond,
		CycleRunTimes:     req.CycleRunTimes,
		Enabled:           req.Enabled,
		RunParams:         req.RunParams,
	}
	task.Id = req.Id
	if task.Id == 0 {
		task.Id = int64(utils.GenerateID())
	}
	return task
}
func listOptions(req QueryJobRequest) queryOptions {
	return queryOptions{JobName: req.JobName, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func (service *Service) validateUnique(ctx context.Context, task Job) error {
	exists, err := service.repository.ExistsByName(ctx, task.JobName, task.Id)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("任务名称=>" + task.JobName + "=>已存在!")
	}
	exists, err = service.repository.ExistsByExecution(ctx, task.AssemblyName, task.ClassName, task.Id)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("执行类=>" + task.ClassName + "=>已存在!")
	}
	return nil
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }

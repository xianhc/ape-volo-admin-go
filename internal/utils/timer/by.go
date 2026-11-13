package timer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	Cron        = 1
	Simple      = 2
	TaskTimeout = 5 * time.Minute
)

// TriggerState 表示对外暴露的调度触发器状态。
//
// 数值与管理前端及参考调度器保持一致。
type TriggerState int32

const (
	// TriggerStateNone 表示当前没有调度计划。
	//
	// 未注册、已停止或已删除的任务返回此状态。
	TriggerStateNone TriggerState = iota
	// TriggerStateNormal 表示调度计划正在等待或触发执行。
	//
	// 该状态不表示任务函数此刻正在运行。
	TriggerStateNormal
	// TriggerStatePaused 表示调度计划已暂停。
	//
	// 恢复后继续按照原配置计算后续触发时间。
	TriggerStatePaused
	// TriggerStateComplete 表示有限次数或时间窗口已经结束。
	//
	// 完成的计划仍保留状态，重新调度会替换它。
	TriggerStateComplete
	// TriggerStateError 表示调度器无法继续触发任务。
	//
	// 当前进程内实现保留该契约值供状态映射使用。
	TriggerStateError
	// TriggerStateBlocked 表示触发器暂时被调度器阻塞。
	//
	// 当前进程内实现通过互斥跳过重叠执行，不产生该状态。
	TriggerStateBlocked
)

// Timer 管理本进程内的任务注册、互斥执行与退出等待。
type Timer interface {
	AddTaskByFunc(ScheduledJob, func(context.Context) error) (cron.EntryID, error)
	GetTaskStatus(int64) TriggerState
	FindTaskStatus(int64) (*taskStatus, bool)
	StartTask(int64) bool
	StopTask(int64) bool
	DeleteByID(int64) bool
	Shutdown(context.Context) error
	Close()
}

type timer struct {
	mu       sync.Mutex
	taskList map[int64]*taskStatus
	running  map[int64]bool
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	closed   bool
}

type taskStatus struct {
	Task     *cron.Cron
	State    TriggerState
	EntryID  int
	Runs     int32
	endTimer *time.Timer
}

// Validate 在修改持久化配置前校验调度参数。
func Validate(task ScheduledJob) error {
	if task.Id <= 0 {
		return errors.New("调度作业必须先保存并取得 ID")
	}
	switch task.TriggerType {
	case Cron:
		if _, err := cron.ParseStandard(task.Cron); err != nil {
			return err
		}
	case Simple:
		if task.IntervalSecond == nil || *task.IntervalSecond <= 0 {
			return errors.New("任务执行间隔必须大于零")
		}
	default:
		return errors.New("不支持的任务触发器类型")
	}
	start := time.Now()
	if task.StartTime != nil {
		start = *task.StartTime
	}
	if task.EndTime != nil && !task.EndTime.After(start) {
		return errors.New("作业结束时间必须晚于开始时间")
	}
	return nil
}

// AddTaskByFunc 按稳定主键幂等替换任务，且同一主键的执行体不会重叠运行。
func (t *timer) AddTaskByFunc(task ScheduledJob, fn func(context.Context) error) (cron.EntryID, error) {
	if err := Validate(task); err != nil {
		return 0, err
	}
	if fn == nil {
		return 0, errors.New("任务执行函数不能为空")
	}
	var schedule cron.Schedule
	var err error
	if task.TriggerType == Simple {
		schedule = cron.Every(time.Duration(*task.IntervalSecond) * time.Second)
	} else {
		schedule, err = cron.ParseStandard(task.Cron)
		if err != nil {
			return 0, err
		}
	}
	schedule = boundedSchedule{Schedule: schedule, StartTime: task.StartTime, EndTime: task.EndTime}
	clock := cron.New()
	status := &taskStatus{Task: clock, State: TriggerStateNormal}
	entry := clock.Schedule(schedule, cron.FuncJob(func() { t.run(task.Id, status, task, fn) }))
	status.EntryID = int(entry)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		clock.Stop()
		return 0, errors.New("任务调度器已关闭")
	}
	if old := t.taskList[task.Id]; old != nil {
		stopStatus(old)
	}
	t.taskList[task.Id] = status
	if task.EndTime != nil {
		status.endTimer = time.AfterFunc(time.Until(*task.EndTime), func() { t.complete(task.Id, status) })
	}
	clock.Start()
	return entry, nil
}

func (t *timer) run(id int64, status *taskStatus, task ScheduledJob, fn func(context.Context) error) {
	t.mu.Lock()
	if t.closed || t.taskList[id] != status || status.State != TriggerStateNormal || t.running[id] {
		t.mu.Unlock()
		return
	}
	t.running[id] = true
	t.wg.Add(1)
	t.mu.Unlock()
	ctx, cancel := context.WithTimeout(t.ctx, TaskTimeout)
	err := fn(ctx)
	cancel()
	t.mu.Lock()
	delete(t.running, id)
	if t.taskList[id] == status {
		status.Runs++
		if err != nil && task.PauseAfterFailure {
			status.Task.Stop()
			status.State = TriggerStatePaused
		} else if task.TriggerType == Simple && task.CycleRunTimes != nil && *task.CycleRunTimes > 0 && status.Runs >= *task.CycleRunTimes {
			status.Task.Stop()
			status.State = TriggerStateComplete
		}
	}
	t.mu.Unlock()
	t.wg.Done()
}

func (t *timer) complete(id int64, status *taskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskList[id] == status {
		status.Task.Stop()
		status.State = TriggerStateComplete
	}
}

func (t *timer) GetTaskStatus(id int64) TriggerState {
	status, ok := t.FindTaskStatus(id)
	if !ok {
		return TriggerStateNone
	}
	return status.State
}

// FindTaskStatus 按稳定主键返回状态副本，调用方不与调度变更共享可变字段。
func (t *timer) FindTaskStatus(id int64) (*taskStatus, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	status, ok := t.taskList[id]
	if !ok {
		return nil, false
	}
	copy := *status
	return &copy, true
}

func (t *timer) StartTask(id int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if status := t.taskList[id]; status != nil && !t.closed && status.State == TriggerStatePaused {
		status.Task.Start()
		status.State = TriggerStateNormal
		return true
	}
	return false
}

func (t *timer) StopTask(id int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if status := t.taskList[id]; status != nil && !t.closed && status.State == TriggerStateNormal {
		status.Task.Stop()
		status.State = TriggerStatePaused
		return true
	}
	return false
}

// DeleteByID 按稳定主键移除任务，避免改名后遗留旧的调度实例。
func (t *timer) DeleteByID(id int64) bool {
	if id == 0 {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	if status := t.taskList[id]; status != nil {
		stopStatus(status)
		delete(t.taskList, id)
	}
	return true
}

// Shutdown 停止调度、取消在途任务，并在截止时间内等待退出。
func (t *timer) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	t.closed = true
	t.cancel()
	for _, status := range t.taskList {
		stopStatus(status)
	}
	t.mu.Unlock()
	done := make(chan struct{})
	go func() { t.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *timer) Close() { _ = t.Shutdown(context.Background()) }

func NewTimerTask() Timer {
	ctx, cancel := context.WithCancel(context.Background())
	return &timer{taskList: make(map[int64]*taskStatus), running: make(map[int64]bool), ctx: ctx, cancel: cancel}
}

type ScheduledJob struct {
	Id                                               int64
	TriggerType                                      int32
	JobName, JobGroup, Cron, ClassName, AssemblyName string
	IntervalSecond                                   *int32
	CycleRunTimes                                    *int32
	StartTime, EndTime                               *time.Time
	PauseAfterFailure                                bool
	Enabled                                          bool
}

type boundedSchedule struct {
	cron.Schedule
	StartTime, EndTime *time.Time
}

func (schedule boundedSchedule) Next(after time.Time) time.Time {
	if schedule.StartTime != nil && after.Before(*schedule.StartTime) {
		after = schedule.StartTime.Add(-time.Nanosecond)
	}
	next := schedule.Schedule.Next(after)
	if schedule.EndTime != nil && next.After(*schedule.EndTime) {
		return time.Time{}
	}
	return next
}

func stopStatus(status *taskStatus) {
	status.Task.Stop()
	if status.endTimer != nil {
		status.endTimer.Stop()
	}
}

package timer

import (
	"fmt"
	"github.com/robfig/cron/v3"

	"sync"
	"time"
)

const (
	Cron   = 1
	Simple = 0
)

type Timer interface {
	AddTaskByFunc(taskJob TaskJob, fn func()) (cron.EntryID, error)
	GetTaskStatus(taskName string) (state string)
	FindTaskStatus(taskName string) (*taskStatus, bool)
	StartTask(taskName string)
	StopTask(taskName string)
	Remove(taskName string)
	Delete(taskName string)
	Close()
}

// timer 定时任务管理
type timer struct {
	taskList map[string]*taskStatus
	sync.Mutex
}

type taskStatus struct {
	Task    *cron.Cron
	State   string
	EntryID int
}

// AddTaskByFunc 通过函数的方法添加任务
func (t *timer) AddTaskByFunc(taskJob TaskJob, fn func()) (entryID cron.EntryID, err error) {
	t.Lock()
	defer func() {
		t.Unlock()
	}()

	if _, ok := t.taskList[taskJob.TaskName]; !ok {
		//t.taskList[taskJob.TaskName].Task = cron.New() //cron.WithSeconds()
		t.taskList[taskJob.TaskName] = &taskStatus{
			Task: cron.New(),
		}
	}
	//_, err := cron.ParseStandard(taskJob.Cron) 维护时已检验 这里只按类型判断
	if taskJob.TriggerType == Cron {
		entryID, err = t.taskList[taskJob.TaskName].Task.AddFunc(taskJob.Cron, fn)
	} else if taskJob.TriggerType == Simple {
		cronExpr := fmt.Sprintf("@every %ds", *taskJob.IntervalSecond)
		entryID, err = t.taskList[taskJob.TaskName].Task.AddFunc(cronExpr, fn)
	} else {
		entryID, err = t.taskList[taskJob.TaskName].Task.AddFunc("@every 3600s", fn)
	}
	if err == nil {
		t.taskList[taskJob.TaskName].Task.Start()
		t.taskList[taskJob.TaskName].State = "运行中"
		t.taskList[taskJob.TaskName].EntryID = int(entryID)
	}

	return entryID, err
}

func (t *timer) GetTaskStatus(taskName string) string {
	t.Lock()
	defer t.Unlock()
	if v, ok := t.taskList[taskName]; ok {
		if v.State == "" {
			return "未执行"
		}
		return v.State
	}
	return "未执行"
}

// FindTaskStatus 获取对应taskName的cron 可能会为空
func (t *timer) FindTaskStatus(taskName string) (*taskStatus, bool) {
	t.Lock()
	defer t.Unlock()
	v, ok := t.taskList[taskName]
	return v, ok
}

// StartTask 开始任务
func (t *timer) StartTask(taskName string) {
	t.Lock()
	defer t.Unlock()
	if v, ok := t.taskList[taskName]; ok {
		v.Task.Start()
		v.State = "运行中"
	}
}

// StopTask 停止任务
func (t *timer) StopTask(taskName string) {
	t.Lock()
	defer t.Unlock()
	if v, ok := t.taskList[taskName]; ok {
		v.Task.Stop()
		v.State = "暂停"
	}
}

// Remove 从taskName 删除指定任务
func (t *timer) Remove(taskName string) {
	t.Lock()
	defer t.Unlock()
	if v, ok := t.taskList[taskName]; ok {
		if v.Task != nil {
			v.Task.Remove(cron.EntryID(v.EntryID))
		}
	}
}

// Delete 删除Map
func (t *timer) Delete(taskName string) {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.taskList[taskName]; ok {
		delete(t.taskList, taskName)
	}
}

// Close 释放资源
func (t *timer) Close() {
	t.Lock()
	defer t.Unlock()
	for _, v := range t.taskList {
		v.Task.Stop()
	}
}

func NewTimerTask() Timer {
	return &timer{taskList: make(map[string]*taskStatus)}
}

type TaskJob struct {
	Id             int64      //主键编码
	TriggerType    int32      //触发器类型（0、simple 1、cron）
	TaskName       string     // 任务名称
	TaskGroup      string     // 任务分组
	Cron           string     // 表达式
	ClassName      string     // 任务所在类
	AssemblyName   string     // 程序集名称
	StartTime      *time.Time // 开始时间
	EndTime        *time.Time // 结束时间
	IntervalSecond *int32     // 执行间隔时间, 秒为单位
	IsEnable       bool       //是否启动
}

package job

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/scheduledjoblog"
	"go.uber.org/zap"
)

// FuncType 定义函数类型
type FuncType func(context.Context) error

// GlobalFuncs 全局函数映射
var GlobalFuncs = map[string]FuncType{}

func RegisterTaskFuncs() {
	// 只注册明确允许执行的任务函数。
	funcs := []FuncType{
		PrintTimeJob,
		SendEmailJob,
	}

	// 将函数名及对应的函数添加到 GlobalFuncs 中
	for _, fn := range funcs {
		name := getFunctionName(fn)
		GlobalFuncs[name] = fn
	}
}

// 获取函数名
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// Prepare 为启动注册与运行时注册提供同一执行入口。
func Prepare(task Definition) (FuncType, error) {
	fn := GlobalFuncs[task.AssemblyName+task.ClassName]
	if fn == nil {
		return nil, fmt.Errorf("作业函数不存在,请检查")
	}
	return func(ctx context.Context) (runErr error) {
		start := time.Now()
		defer func() {
			if r := recover(); r != nil {
				runErr = fmt.Errorf("任务 panic: %v", r)
			}
			if runErr == nil {
				runErr = ctx.Err()
			}
			if runErr != nil {
				global.Logger.Error("任务执行失败", zap.String("task", task.JobName), zap.Error(runErr))
			}
			(&scheduledjoblog.Service{}).Record(ctx, scheduledjoblog.Execution{
				JobId:        task.Id,
				JobName:      task.JobName,
				JobGroup:     task.JobGroup,
				AssemblyName: task.AssemblyName,
				ClassName:    task.ClassName,
				Cron:         task.Cron,
				RunParams:    task.RunParams,
			}, time.Since(start), runErr)
		}()
		return fn(ctx)
	}, nil
}

package initialize

import (
	"context"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/job"
	jobmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/system/job"
	"go.uber.org/zap"
)

// Timer 加载启用任务，执行入口与运行时修改共用。
func Timer() error {
	job.RegisterTaskFuncs()
	service := jobmodule.Service{}
	tasks, err := service.Definitions(context.Background())
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Enabled {
			if err := service.RegisterRuntime(task); err != nil {
				global.Logger.Error("任务配置不可执行，已跳过", zap.String("task", task.JobName), zap.Error(err))
			}
		}
	}
	return nil
}

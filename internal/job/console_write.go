package job

import (
	"context"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
	"go.uber.org/zap"
)

func PrintTimeJob(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	global.Logger.Info("定时任务输出当前时间", zap.Any("time", ext.GetCurrentTime()))
	return nil
}

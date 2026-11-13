package initialize

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/exceptionlog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/operatelog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/scheduledjoblog"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
	"go.uber.org/zap"
)

// StartLogWriters 在 HTTP 和调度启动前启动三个独立消费者。
//
// 部分启动失败会排空已经启动的队列。
func StartLogWriters() error {
	for _, item := range []struct {
		name  string
		start func(logqueue.Options) error
	}{
		{"operate", operatelog.StartWriter}, {"exception", exceptionlog.StartWriter}, {"scheduled-job", scheduledjoblog.StartWriter},
	} {
		name := item.name
		var lastFull, lastClosed, lastFailed, lastUnknown int64
		summary := func(s logqueue.Stats) {
			fields := []zap.Field{zap.String("type", name), zap.Any("stats", s)}
			if s.DroppedFull > lastFull || s.DroppedClosed > lastClosed || s.WriteFailed > lastFailed || s.CommitUnknown > lastUnknown || s.Remaining > 0 {
				global.Logger.Warn("日志队列丢弃或写入异常摘要", fields...)
			} else if s.Length == 0 && s.InFlight == 0 {
				global.Logger.Debug("日志队列状态", fields...)
			}
			lastFull, lastClosed, lastFailed, lastUnknown = s.DroppedFull, s.DroppedClosed, s.WriteFailed, s.CommitUnknown
		}
		if err := item.start(logqueue.Options{Capacity: global.Config.LogQueue.Capacity, BatchSize: global.Config.LogQueue.BatchSize, Summary: summary}); err != nil {
			return errors.Join(err, StopLogWriters())
		}
	}
	return nil
}

// StopLogWriters 使用一个独立总期限并行排空三类队列。
//
// 所有消费者退出后返回，调用方此后才能关闭数据库。
func StopLogWriters() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(global.Config.LogQueue.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	results := make(chan error, 3)
	for _, stop := range []func(context.Context) error{operatelog.StopWriter, exceptionlog.StopWriter, scheduledjoblog.StopWriter} {
		wg.Go(func() { results <- stop(ctx) })
	}
	wg.Wait()
	close(results)
	var err error
	for result := range results {
		err = errors.Join(err, result)
	}
	global.Logger.Info("日志队列停止", zap.Any("operate", operatelog.WriterStats()), zap.Any("exception", exceptionlog.WriterStats()), zap.Any("scheduled_job", scheduledjoblog.WriterStats()))
	return err
}

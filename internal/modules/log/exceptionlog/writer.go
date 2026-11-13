package exceptionlog

import (
	"context"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
)

var writer logqueue.Queue[ExceptionLog]

// StartWriter 启动异常日志的独立消费者。
//
// 不与操作日志共享容量。
func StartWriter(options logqueue.Options) error {
	return writer.Start(options, func(ctx context.Context, entries []ExceptionLog) logqueue.Result {
		r := (repository{}).CreateBatch(ctx, entries)
		return logqueue.Result{Persisted: r.Persisted, Failed: r.Failed, Unknown: r.Unknown, Remaining: r.Remaining}
	})
}

// StopWriter 关闭异常日志入口并等待排空。
//
// 截止后取消写入，仍等待消费者释放连接。
func StopWriter(ctx context.Context) error { return writer.Stop(ctx) }

// WriterStats 返回异常日志队列快照。
//
// 不查询数据库，也不影响消费者。
func WriterStats() logqueue.Stats { return writer.Stats() }

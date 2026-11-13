package scheduledjoblog

import (
	"context"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
)

var writer logqueue.Queue[ScheduledJobLog]

// StartWriter 启动作业日志的独立消费者。
//
// 写入失败不改变作业执行结果。
func StartWriter(options logqueue.Options) error {
	return writer.Start(options, func(ctx context.Context, entries []ScheduledJobLog) logqueue.Result {
		r := (repository{}).CreateBatch(ctx, entries)
		return logqueue.Result{Persisted: r.Persisted, Failed: r.Failed, Unknown: r.Unknown, Remaining: r.Remaining}
	})
}

// StopWriter 关闭作业日志入口并等待排空。
//
// 截止后取消写入，仍等待消费者释放连接。
func StopWriter(ctx context.Context) error { return writer.Stop(ctx) }

// WriterStats 返回作业日志队列快照。
//
// 不查询数据库，也不影响消费者。
func WriterStats() logqueue.Stats { return writer.Stats() }

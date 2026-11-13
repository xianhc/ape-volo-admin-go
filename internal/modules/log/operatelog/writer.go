package operatelog

import (
	"context"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/logqueue"
)

var writer logqueue.Queue[OperateLog]

// StartWriter 启动操作日志的独立消费者。
//
// 消费者只通过本模块仓储写库。
func StartWriter(options logqueue.Options) error {
	return writer.Start(options, func(ctx context.Context, entries []OperateLog) logqueue.Result {
		r := (repository{}).CreateBatch(ctx, entries)
		return logqueue.Result{Persisted: r.Persisted, Failed: r.Failed, Unknown: r.Unknown, Remaining: r.Remaining}
	})
}

// StopWriter 关闭操作日志入口并等待排空。
//
// 截止后取消写入，仍等待消费者释放连接。
func StopWriter(ctx context.Context) error { return writer.Stop(ctx) }

// WriterStats 返回操作日志队列快照。
//
// 不查询数据库，也不影响消费者。
func WriterStats() logqueue.Stats { return writer.Stats() }

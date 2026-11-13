// Package logqueue 提供有界、非阻塞、单消费者的内存日志队列。
package logqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// ErrRejected 表示生产入口未接收记录。
//
// 调用方不能据此同步回写业务数据库。
var ErrRejected = errors.New("日志队列已关闭、未启动或已满")

// ClonePointer 复制可空值，用于生产端构造独立快照。
//
// T 自身包含的可变引用仍需调用方按字段语义复制。
func ClonePointer[T any](p *T) *T {
	if p == nil {
		return nil
	}
	value := *p
	return &value
}

// Result 描述一个消费者批次的处理分类。
//
// 各分类数量合计应等于批次条数，Unknown 不等于确定丢失。
type Result struct{ Persisted, Failed, Unknown, Remaining int }

// Stats 提供并发安全的进程内队列计数快照。
//
// 运行期间各计数可能来自相邻时刻，排空结束后可核对总数。
type Stats struct {
	Length                                                                                           int
	Accepted, Persisted, DroppedFull, DroppedClosed, WriteFailed, CommitUnknown, Remaining, InFlight int64
}

// Options 配置单类队列容量及最大批次。
//
// Summary 串行运行，最多每十秒和退出时各汇总一次。
type Options struct {
	Capacity, BatchSize int
	Summary             func(Stats)
}

// Queue 提供多生产者、单消费者的非阻塞有界队列。
//
// 零值拒绝写入并计数，Start 不继承请求或应用取消。
type Queue[T any] struct {
	mu                                                                          sync.RWMutex
	items                                                                       chan T
	done                                                                        chan struct{}
	cancel                                                                      context.CancelFunc
	closed                                                                      bool
	accepted, persisted, full, closedDrop, failed, unknown, remaining, inFlight atomic.Int64
}

// Start 启动使用独立 context 的消费者。
//
// 只有排空超时才取消，先前消费者完全结束后可再次启动。
func (q *Queue[T]) Start(options Options, write func(context.Context, []T) Result) error {
	if options.Capacity < 1 || options.BatchSize < 1 || options.BatchSize > 200 || write == nil {
		return errors.New("日志队列参数无效")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	restarting := q.items != nil
	if q.items != nil {
		select {
		case <-q.done:
		default:
			return errors.New("日志队列已启动")
		}
	}
	q.items = make(chan T, options.Capacity)
	q.done = make(chan struct{})
	q.closed = false
	for _, counter := range []*atomic.Int64{&q.accepted, &q.persisted, &q.full, &q.closedDrop, &q.failed, &q.unknown, &q.remaining, &q.inFlight} {
		if counter == &q.closedDrop && !restarting {
			continue
		}
		counter.Store(0)
	}
	ctx, cancel := context.WithCancel(context.Background())
	q.cancel = cancel
	go q.consume(ctx, q.items, q.done, options, write)
	return nil
}

// TryEnqueue 尝试发送一条完整快照。
//
// 发送与关闭共享同步门，容量耗尽时立即拒绝。
func (q *Queue[T]) TryEnqueue(value T) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.items == nil || q.closed {
		q.closedDrop.Add(1)
		return false
	}
	select {
	case q.items <- value:
		q.accepted.Add(1)
		return true
	default:
		q.full.Add(1)
		return false
	}
}

// Close 关闭生产入口，保留已接收记录供消费者排空。
//
// 可重复调用，并发发送只会被接收或拒绝，不会发送到已关闭通道。
func (q *Queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closeLocked()
}

func (q *Queue[T]) closeLocked() {
	if q.items != nil && !q.closed {
		q.closed = true
		close(q.items)
	}
}

// Stop 关闭入口并在截止时间内尝试排空。
//
// 必须等消费者返回后再允许调用方关闭连接，即使驱动取消晚于期限。
func (q *Queue[T]) Stop(ctx context.Context) error {
	q.mu.Lock()
	q.closeLocked()
	done, cancel := q.done, q.cancel
	q.mu.Unlock()
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	default:
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		cancel()
		<-done
		return ctx.Err()
	}
}

// Stats 返回当前队列长度及处理计数。
//
// 未启动队列也包含 closed 丢弃数量。
func (q *Queue[T]) Stats() Stats {
	q.mu.RLock()
	length := len(q.items)
	q.mu.RUnlock()
	return Stats{Length: length, Accepted: q.accepted.Load(), Persisted: q.persisted.Load(), DroppedFull: q.full.Load(), DroppedClosed: q.closedDrop.Load(), WriteFailed: q.failed.Load(), CommitUnknown: q.unknown.Load(), Remaining: q.remaining.Load(), InFlight: q.inFlight.Load()}
}

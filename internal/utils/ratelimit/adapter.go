package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ulule/limiter/v3"
)

type adapter struct {
	mu       sync.RWMutex
	instance *limiter.Limiter
	limit    int64
	window   time.Duration
	name     string
}

func normalize(limit int64, window time.Duration) (int64, time.Duration) {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Second
	}
	return limit, window
}

func newAdapter(store limiter.Store, limit int64, window time.Duration, name string) Limiter {
	return &adapter{
		instance: limiter.New(store, limiter.Rate{Limit: limit, Period: window}),
		limit:    limit, window: window, name: name,
	}
}

func (l *adapter) Name() string          { return l.name }
func (l *adapter) Limit() int64          { return l.limit }
func (l *adapter) Window() time.Duration { return l.window }

func (l *adapter) Allow(ctx context.Context, key string) (Result, error) {
	l.mu.RLock()
	instance := l.instance
	l.mu.RUnlock()
	if instance == nil {
		return Result{}, fmt.Errorf("ratelimit: 限流器已关闭")
	}
	state, err := instance.Get(ctx, key)
	if err != nil {
		return Result{}, fmt.Errorf("ratelimit: %s 判定失败: %w", l.name, err)
	}
	result := Result{Allowed: !state.Reached, Limit: state.Limit, Remaining: max(state.Remaining, 0)}
	if state.Reached {
		result.RetryAfter = time.Until(time.Unix(state.Reset, 0))
		// 重置时间只有秒精度，同一秒内尚未结束的窗口仍需给出正等待时间。
		if result.RetryAfter <= 0 {
			result.RetryAfter = min(l.window, time.Second)
		}
	}
	return result, nil
}

func (l *adapter) Close() error {
	l.mu.Lock()
	l.instance = nil
	l.mu.Unlock()
	return nil
}

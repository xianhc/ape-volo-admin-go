package ratelimit

import (
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// NewMemoryLimiter 创建进程内固定窗口限流器。
//
// Close 释放存储引用，清理协程在存储不再被引用后由 GC 最终停止。
func NewMemoryLimiter(limit int64, window time.Duration) Limiter {
	limit, window = normalize(limit, window)
	store := memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix: "ratelimit", CleanUpInterval: max(window, time.Second),
	})
	return newAdapter(store, limit, window, "memory")
}

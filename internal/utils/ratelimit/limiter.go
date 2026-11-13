// Package ratelimit 提供限流抽象。
//
// 两套实现：
//   - 内存固定窗口：单实例部署，无外部存储；
//   - Redis 固定窗口：多实例部署，全局限流共享计数。
//
// 由 system.use-redis 决定使用哪一套，见 ADR-002。
//
// 选择固定窗口而非令牌桶：本项目的限流只用于登录、验证码这类
// 防爆破场景，固定窗口实现简单、边界行为可预测；令牌桶在窗口
// 边界处的平滑性对本场景没有实际收益，却要维护额外的状态与时钟。
package ratelimit

import (
	"context"
	"time"
)

// Result 是一次限流判定的结果。
type Result struct {
	// Allowed 是否放行
	Allowed bool
	// Limit 窗口内的配额
	Limit int64
	// Remaining 窗口内剩余可用次数
	Remaining int64
	// RetryAfter 建议的重试等待时间，仅在 Allowed=false 时有意义
	RetryAfter time.Duration
}

// Limiter 限流器抽象。
//
// key 由调用方构造（通常是 "路径:客户端IP"），实现不感知业务语义。
type Limiter interface {
	Allow(ctx context.Context, key string) (Result, error)
	// Limit 返回窗口内配额，用于写入响应头
	Limit() int64
	// Window 返回统计窗口长度
	Window() time.Duration
	// Name 返回实现名称，用于启动日志
	Name() string
	// Close 释放存储引用，可重复调用；内存清理协程由 GC 最终停止。
	Close() error
}

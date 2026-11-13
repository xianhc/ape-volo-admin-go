// Package cache 提供统一的缓存抽象。
//
// 设计要点：
//   - 接口以「Go 值」为单位，内部统一做 JSON 编解码，调用方无需关心序列化。
//   - 两套实现：redisStore（多实例共享）与 memoryStore（单实例，进程内）。
//   - 由 system.use-redis 决定注入哪一套，见 ADR-002。
//
// 约定：Get 在 key 不存在时返回 ErrNotFound，调用方应使用 errors.Is 判断，
// 不要依赖具体实现的错误值。
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound 表示缓存键不存在。
var ErrNotFound = errors.New("cache: key not found")

// Store 是缓存后端的统一抽象。
//
// 所有方法都必须是并发安全的。value 为 nil 时 Set 视为删除该键。
type Store interface {
	// Get 读取 key 并反序列化到 dest（dest 必须是非 nil 指针）。
	// key 不存在时返回 ErrNotFound。
	Get(ctx context.Context, key string, dest any) error

	// Set 序列化 value 并写入 key，expiration <= 0 表示永不过期。
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Del 删除一个或多个 key，key 不存在不视为错误。
	Del(ctx context.Context, keys ...string) error

	// Incr 对 key 做原子自增并返回自增后的值。
	// key 不存在时按 0 处理（即结果为 delta）。
	// 已存在的值必须是可解析为整数的字符串，否则返回错误。
	Incr(ctx context.Context, key string, delta int64) (int64, error)

	// CompareExchange 原子地比较并替换：仅当 key 当前值等于 old 时写入 new。
	//
	// old 为 nil 表示要求 key 不存在；new 为 nil 表示删除。
	// 返回是否发生了替换。用于会话轮换等「最多一个并发请求成功」的场景。
	CompareExchange(ctx context.Context, key string, old, new any, expiration time.Duration) (bool, error)

	// Keys 返回匹配 pattern 的所有 key。pattern 支持 * 通配（glob 语义）。
	Keys(ctx context.Context, pattern string) ([]string, error)

	// DeleteByPrefix 删除所有以 prefix 开头的 key，返回删除数量。
	DeleteByPrefix(ctx context.Context, prefix string) (int64, error)

	// Close 释放后端资源。memoryStore 的 Close 会停止过期清理协程。
	Close() error

	// Name 返回实现名称，用于启动日志与健康检查。
	Name() string
}

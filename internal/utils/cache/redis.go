package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisStore 是基于 Redis 的缓存实现，用于多实例部署。
//
// Keys / DeleteByPrefix 一律使用 SCAN 而非 KEYS：
// KEYS 在大 key 空间下会阻塞 Redis 单线程，属于生产事故常见诱因。
type redisStore struct {
	client *redis.Client
}

// NewRedisStore 包装一个已建立的 Redis 客户端。
func NewRedisStore(client *redis.Client) Store {
	return &redisStore{client: client}
}

func (s *redisStore) Name() string { return "redis" }

func (s *redisStore) Get(ctx context.Context, key string, dest any) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return fmt.Errorf("cache: redis get %q: %w", key, err)
	}
	return decode(data, dest)
}

func (s *redisStore) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if value == nil {
		return s.Del(ctx, key)
	}
	data, err := encode(value)
	if err != nil {
		return err
	}
	if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("cache: redis set %q: %w", key, err)
	}
	return nil
}

func (s *redisStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache: redis del: %w", err)
	}
	return nil
}

func (s *redisStore) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	n, err := s.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("cache: redis incr %q: %w", key, err)
	}
	return n, nil
}

// compareExchangeScript 用 Lua 保证「比较 + 替换」在 Redis 侧原子完成。
//
// ARGV[1] 期望值（空串表示要求键不存在），ARGV[2] 新值（空串表示删除），
// ARGV[3] TTL 毫秒（0 表示不过期）。
var compareExchangeScript = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if ARGV[1] == '' then
  if current then return 0 end
else
  if current ~= ARGV[1] then return 0 end
end
if ARGV[2] == '' then
  redis.call('DEL', KEYS[1])
else
  if tonumber(ARGV[3]) > 0 then
    redis.call('SET', KEYS[1], ARGV[2], 'PX', ARGV[3])
  else
    redis.call('SET', KEYS[1], ARGV[2])
  end
end
return 1
`)

func (s *redisStore) CompareExchange(ctx context.Context, key string, old, new any, expiration time.Duration) (bool, error) {
	var oldStr, newStr string
	if old != nil {
		data, err := encode(old)
		if err != nil {
			return false, err
		}
		oldStr = string(data)
	}
	if new != nil {
		data, err := encode(new)
		if err != nil {
			return false, err
		}
		newStr = string(data)
	}

	ttlMs := int64(0)
	if expiration > 0 {
		ttlMs = expiration.Milliseconds()
	}

	result, err := compareExchangeScript.Run(ctx, s.client, []string{key}, oldStr, newStr, ttlMs).Int64()
	if err != nil {
		return false, fmt.Errorf("cache: redis compare-exchange %q: %w", key, err)
	}
	return result == 1, nil
}

func (s *redisStore) Keys(ctx context.Context, pattern string) ([]string, error) {
	return s.scan(ctx, pattern)
}

func (s *redisStore) DeleteByPrefix(ctx context.Context, prefix string) (int64, error) {
	keys, err := s.scan(ctx, prefix+"*")
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return 0, fmt.Errorf("cache: redis delete by prefix %q: %w", prefix, err)
	}
	return int64(len(keys)), nil
}

func (s *redisStore) Close() error { return s.client.Close() }

// scan 用游标遍历匹配的 key，避免 KEYS 阻塞。
func (s *redisStore) scan(ctx context.Context, pattern string) ([]string, error) {
	const batch = 500

	var (
		cursor uint64
		keys   []string
	)
	for {
		page, next, err := s.client.Scan(ctx, cursor, pattern, batch).Result()
		if err != nil {
			return nil, fmt.Errorf("cache: redis scan %q: %w", pattern, err)
		}
		keys = append(keys, page...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

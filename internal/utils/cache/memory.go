package cache

import (
	"bytes"
	"context"
	"strconv"
	"sync"
	"time"
)

// memoryStore 是进程内缓存实现，用于 system.use-redis = false 的单实例部署。
//
// 语义与 redisStore 保持一致（值统一 JSON 编码），因此两种模式可以无缝切换。
// 过期采用「惰性删除 + 周期性清理」：读时判断是否过期，另有 janitor 定时回收。
type memoryStore struct {
	mu     sync.RWMutex
	items  map[string]memoryItem
	stop   chan struct{}
	closed sync.Once
}

type memoryItem struct {
	data     []byte
	expireAt time.Time // 零值表示永不过期
}

func (i memoryItem) expired(now time.Time) bool {
	return !i.expireAt.IsZero() && now.After(i.expireAt)
}

// NewMemoryStore 创建内存缓存实现。
//
// sweepInterval 为过期清理周期，<= 0 时使用 1 分钟默认值。
func NewMemoryStore(sweepInterval time.Duration) Store {
	if sweepInterval <= 0 {
		sweepInterval = time.Minute
	}
	s := &memoryStore{
		items: make(map[string]memoryItem),
		stop:  make(chan struct{}),
	}
	go s.janitor(sweepInterval)
	return s
}

func (s *memoryStore) Name() string { return "memory" }

func (s *memoryStore) Get(_ context.Context, key string, dest any) error {
	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()

	if !ok {
		return ErrNotFound
	}
	if item.expired(time.Now()) {
		s.mu.Lock()
		// 二次确认，避免并发下误删刚被写入的新值
		if cur, exist := s.items[key]; exist && cur.expired(time.Now()) {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return ErrNotFound
	}
	return decode(item.data, dest)
}

func (s *memoryStore) Set(_ context.Context, key string, value any, expiration time.Duration) error {
	if value == nil {
		return s.Del(context.Background(), key)
	}
	data, err := encode(value)
	if err != nil {
		return err
	}
	var expireAt time.Time
	if expiration > 0 {
		expireAt = time.Now().Add(expiration)
	}

	s.mu.Lock()
	s.items[key] = memoryItem{data: data, expireAt: expireAt}
	s.mu.Unlock()
	return nil
}

func (s *memoryStore) Del(_ context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	s.mu.Lock()
	for _, k := range keys {
		delete(s.items, k)
	}
	s.mu.Unlock()
	return nil
}

func (s *memoryStore) Incr(_ context.Context, key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	item, ok := s.items[key]
	var current int64
	var expireAt time.Time

	if ok && !item.expired(now) {
		n, err := decodeInt(item.data)
		if err != nil {
			return 0, err
		}
		current = n
		expireAt = item.expireAt // 保留原 TTL，与 Redis INCR 行为一致
	}

	next := current + delta
	s.items[key] = memoryItem{
		data:     []byte(strconv.FormatInt(next, 10)),
		expireAt: expireAt,
	}
	return next, nil
}

func (s *memoryStore) CompareExchange(_ context.Context, key string, old, new any, expiration time.Duration) (bool, error) {
	var oldData []byte
	if old != nil {
		var err error
		if oldData, err = encode(old); err != nil {
			return false, err
		}
	}

	var newData []byte
	if new != nil {
		var err error
		if newData, err = encode(new); err != nil {
			return false, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.items[key]
	if current.expired(time.Now()) {
		exists = false
	}

	// old 为 nil 时要求键不存在
	if old == nil {
		if exists {
			return false, nil
		}
	} else {
		if !exists || !bytes.Equal(current.data, oldData) {
			return false, nil
		}
	}

	if new == nil {
		delete(s.items, key)
		return true, nil
	}

	var expireAt time.Time
	if expiration > 0 {
		expireAt = time.Now().Add(expiration)
	}
	s.items[key] = memoryItem{data: newData, expireAt: expireAt}
	return true, nil
}

func (s *memoryStore) Keys(_ context.Context, pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	keys := make([]string, 0, len(s.items))
	for k, item := range s.items {
		if item.expired(now) {
			continue
		}
		if matchPattern(pattern, k) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (s *memoryStore) DeleteByPrefix(_ context.Context, prefix string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int64
	for k := range s.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(s.items, k)
			n++
		}
	}
	return n, nil
}

func (s *memoryStore) Close() error {
	s.closed.Do(func() { close(s.stop) })
	return nil
}

// janitor 周期性清理过期键，避免长期运行后内存只增不减。
func (s *memoryStore) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.mu.Lock()
			for k, item := range s.items {
				if item.expired(now) {
					delete(s.items, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

// matchPattern 实现 glob 匹配，支持 * 与 ?。
//
// 不直接用 path.Match：缓存 key 中可能包含 [ ] 等字符，
// path.Match 会把它们当作字符类语法从而匹配失败。
func matchPattern(pattern, s string) bool {
	// 双指针回溯法，避免递归爆栈
	var (
		p, i     int
		starIdx  = -1
		matchIdx int
	)
	for i < len(s) {
		switch {
		case p < len(pattern) && (pattern[p] == '?' || pattern[p] == s[i]):
			p++
			i++
		case p < len(pattern) && pattern[p] == '*':
			starIdx = p
			matchIdx = i
			p++
		case starIdx != -1:
			p = starIdx + 1
			matchIdx++
			i = matchIdx
		default:
			return false
		}
	}
	for p < len(pattern) && pattern[p] == '*' {
		p++
	}
	return p == len(pattern)
}

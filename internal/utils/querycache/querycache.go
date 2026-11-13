// Package querycache 提供「先读缓存、未命中回源」的查询缓存。
//
// 为什么不放在 utils/cache 里：缓存开关 aop.cache 与缓存实例 global.Cache 都在 global 包，
// 而 global 又依赖 utils/cache，把 GetOrLoad 写进 cache 会形成导入环。
// 这里位于 global 之上，可以同时引用配置与缓存实例。
package querycache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
	"go.uber.org/zap"
)

// coherence 只协调本进程的回源与失效，不存储业务缓存值。
var coherence = struct {
	sync.RWMutex
	epoch   uint64
	blocked map[string]bool
}{blocked: make(map[string]bool)}

func blocked(key string) bool {
	for prefix := range coherence.blocked {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// Invalidate 失效查询缓存，并阻止此前开始的回源重新写入旧值。
// 失效失败的前缀暂时绕过查询缓存，直到下次成功失效。
func Invalidate(ctx context.Context, prefixes ...string) {
	if global.Cache == nil {
		return
	}
	coherence.Lock()
	defer coherence.Unlock()
	coherence.epoch++
	for _, prefix := range prefixes {
		if _, err := global.Cache.DeleteByPrefix(ctx, prefix); err != nil {
			coherence.blocked[prefix] = true
			global.Logger.Warn("查询缓存失效失败，该前缀暂时回源数据库", zap.String("prefix", prefix), zap.Error(err))
		} else {
			delete(coherence.blocked, prefix)
		}
	}
}

// Key 保留可读前缀及简单参数，复杂参数序列化后计算 16 位 MD5 摘要。
// 无法序列化的参数返回空键，GetOrLoad 将绕过缓存。
func Key(prefix string, arg any) string {
	value := reflect.ValueOf(arg)
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		value = value.Elem()
	}
	if !value.IsValid() {
		return prefix
	}
	arg = value.Interface()
	if date, ok := arg.(time.Time); ok {
		return prefix + date.Format(time.RFC3339Nano)
	}
	switch value.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return prefix + fmt.Sprint(arg)
	default:
		data, err := json.Marshal(arg)
		if err != nil {
			return ""
		}
		return prefix + utils.MD5(string(data))[8:24]
	}
}

// GetOrLoad 先读缓存，未命中时调用 load 回源并回填。
//
// 行为：
//   - aop.cache 关闭时完全透传，不读写缓存；
//   - 读缓存失败（后端抖动）只记日志并继续回源，不阻断业务；
//   - 回填失败同样只记日志。
//
// 调用方通过 load 闭包提供回源操作。
func GetOrLoad(ctx context.Context, key string, dest any, expiration time.Duration, load func() error) error {
	return getOrLoad(ctx, key, dest, expiration, false, load)
}

// GetOrLoadAlways 使用共享查询缓存，但不受 aop.cache 开关控制。
//
// 仅用于必须跨请求复用的用户菜单；事务绕过、失效保护和缓存故障回源语义与 GetOrLoad 相同。
func GetOrLoadAlways(ctx context.Context, key string, dest any, expiration time.Duration, load func() error) error {
	return getOrLoad(ctx, key, dest, expiration, true, load)
}

func getOrLoad(ctx context.Context, key string, dest any, expiration time.Duration, ignoreSwitch bool, load func() error) error {
	if key == "" || (!ignoreSwitch && !global.Config.Aop.Cache) || global.Cache == nil || (utils.TxFromContext(ctx) != nil && !utils.TxSkipped(ctx)) {
		return load()
	}

	coherence.RLock()
	if blocked(key) {
		coherence.RUnlock()
		return load()
	}
	epoch := coherence.epoch
	err := global.Cache.Get(ctx, key, dest)
	coherence.RUnlock()
	if err == nil {
		return nil
	}
	if !errors.Is(err, cache.ErrNotFound) {
		global.Logger.Warn("缓存读取失败，回源数据库", zap.String("key", key), zap.Error(err))
	}

	if err = load(); err != nil {
		return err
	}

	coherence.RLock()
	defer coherence.RUnlock()
	if epoch != coherence.epoch || blocked(key) {
		return nil
	}
	if err = global.Cache.Set(ctx, key, dest, expiration); err != nil {
		global.Logger.Warn("缓存写入失败", zap.String("key", key), zap.Error(err))
	}
	return nil
}

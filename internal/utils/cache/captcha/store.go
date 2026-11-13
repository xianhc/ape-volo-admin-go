// Package captcha 提供基于 cache.Store 的验证码存储实现。
//
// 与 base64Captcha.DefaultMemStore 的区别：这里的存储跟随 system.use-redis，
// use-redis=true 时验证码写入 Redis（多实例共享，任一节点都能校验），
// use-redis=false 时写入进程内存（单实例）。
package captcha

import (
	"context"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"

	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

// defaultExpiration 是验证码有效期，与原实现保持一致。
const defaultExpiration = 120 * time.Second

// Store 实现 base64Captcha.Store，底层委托给 cache.Store。
type Store struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
}

// NewStore 创建验证码存储。
func NewStore() *Store {
	return &Store{
		Expiration: defaultExpiration,
		PreKey:     cachePrefix.LoginCaptcha,
		Context:    context.Background(),
	}
}

// UseWithCtx 让验证码存储携带请求上下文。
func (s *Store) UseWithCtx(ctx context.Context) base64Captcha.Store {
	s.Context = ctx
	return s
}

func (s *Store) ctx() context.Context {
	if s.Context != nil {
		return s.Context
	}
	return context.Background()
}

// Set 写入验证码。
func (s *Store) Set(id string, value string) error {
	if err := global.Cache.Set(s.ctx(), s.PreKey+id, value, s.Expiration); err != nil {
		global.Logger.Error("验证码写入失败", zap.String("id", id), zap.Error(err))
		return err
	}
	return nil
}

// Get 读取验证码，clear 为 true 时读取后立即删除。
func (s *Store) Get(key string, clear bool) string {
	var value string
	err := global.Cache.Get(s.ctx(), key, &value)
	if err != nil {
		if err != cache.ErrNotFound {
			global.Logger.Error("验证码读取失败", zap.String("key", key), zap.Error(err))
		}
		return ""
	}
	if clear {
		if err := global.Cache.Del(s.ctx(), key); err != nil {
			global.Logger.Error("验证码清理失败", zap.String("key", key), zap.Error(err))
			return ""
		}
	}
	return value
}

// Verify 校验验证码。
func (s *Store) Verify(id, answer string, clear bool) bool {
	key := s.PreKey + id
	return s.Get(key, clear) == answer
}

package captcha

import (
	"context"
	"go-apevolo/global/constants/cachePrefix"
	"time"

	"github.com/mojocn/base64Captcha"
	"go-apevolo/global"
	"go.uber.org/zap"
)

func NewDefaultRedisStore() *RedisStore {
	return &RedisStore{
		Expiration: time.Second * 120,
		PreKey:     cachePrefix.LoginCaptcha,
	}
}

type RedisStore struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
}

func (rs *RedisStore) UseWithCtx(ctx context.Context) base64Captcha.Store {
	rs.Context = ctx
	return rs
}

func (rs *RedisStore) Set(id string, value string) error {
	err := global.Redis.Set(rs.Context, rs.PreKey+id, value, rs.Expiration).Err()
	if err != nil {
		global.Logger.Error("RedisStoreSetError!", zap.Error(err))
		return err
	}
	return nil
}

func (rs *RedisStore) Get(key string, clear bool) string {
	val, err := global.Redis.Get(rs.Context, key).Result()
	if err != nil {
		global.Logger.Error("RedisStoreGetError!", zap.Error(err))
		return ""
	}
	if clear {
		err := global.Redis.Del(rs.Context, key).Err()
		if err != nil {
			global.Logger.Error("RedisStoreClearError!", zap.Error(err))
			return ""
		}
	}
	return val
}

func (rs *RedisStore) Verify(id, answer string, clear bool) bool {
	key := rs.PreKey + id
	v := rs.Get(key, clear)
	return v == answer
}

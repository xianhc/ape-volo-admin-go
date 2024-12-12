package aop

import (
	"errors"
	"fmt"
	"go-apevolo/global"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/redis"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type CacheInterceptor func() []reflect.Value

func CacheAop(cacheConfig CacheConfig, method interface{}, outResult interface{}, args ...interface{}) func() error {
	return func() error {
		cacheKey := ""
		if len(args) > 0 {
			arg, ok := args[0].(string)
			if !ok {
				cacheKey = fmt.Sprint(args[0])
			} else {
				cacheKey = arg
			}
			if err := getCache(cacheConfig.Prefix+utils.MD5(cacheKey), outResult); err != nil {
				if !errors.Is(err, redis.Nil) {
					return err
				}
			} else {
				if cacheConfig.ExpireType == redis.Relative {
					resultValue := reflect.ValueOf(outResult)
					data := resultValue.Interface()
					setCache(cacheConfig.Prefix+utils.MD5(cacheKey), data, cacheConfig.Expiration)
				}
				return nil
			}
		}

		// 调用原始函数，并传递参数
		v := reflect.ValueOf(method)
		var in []reflect.Value
		if len(args) > 0 {
			in = make([]reflect.Value, len(args))
			for i, arg := range args {
				in[i] = reflect.ValueOf(arg)
			}
		}
		// 执行原始方法
		resultValues := v.Call(in)

		// 获取返回的 error
		var err error
		if len(resultValues) > 0 {
			if !resultValues[0].IsNil() {
				err = resultValues[0].Interface().(error) // 获取 error 类型
				global.Logger.Error("cache aop error:", zap.Error(err))
				return err
			}
		}

		resultValue := reflect.ValueOf(outResult)
		data := resultValue.Interface()
		setCache(cacheConfig.Prefix+utils.MD5(cacheKey), data, cacheConfig.Expiration)
		return nil
	}
}

func getCache(key string, result interface{}) error {
	err := redis.Get(key, result)
	if err != nil {
		return err
	}
	return nil
}

func setCache(key string, data interface{}, expiration time.Duration) {
	_ = redis.Set(key, data, expiration)
}

type CacheConfig struct {
	Prefix string //前缀
	//Key        string          //键
	Expiration time.Duration         //时间
	ExpireType redis.CacheExpireType //类型
}

// NewCacheConfig 构造函数
func NewCacheConfig(cachePrefix string, expiration time.Duration, expireType *redis.CacheExpireType) CacheConfig {
	if expiration == 0 {
		expiration = ext.GetTimeDuration(20, ext.Minute) // 设置默认的过期时间
	}

	if expireType == nil {
		expType := redis.Absolute
		expireType = &expType // 设置默认的过期时间
	}

	return CacheConfig{
		Prefix: cachePrefix,
		//Key:        *cacheKey,
		Expiration: expiration,
		ExpireType: *expireType,
	}
}

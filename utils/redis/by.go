package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-apevolo/global"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Get
// @description:获取缓存
// @param: key
// @param: result
// @return: error
func Get(key string, result interface{}) error {
	// 从 Redis 获取字节切片
	serializedData, err := global.Redis.Get(context.Background(), key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			//return err
			return Nil // 返回自定义的错误
		}
		global.Logger.Error("redis get error: "+key, zap.Error(err))
		return err
	}

	err = ext.JsonUnmarshal(serializedData, result)
	if err != nil {
		global.Logger.Error("redis jsonUnmarshal error: "+key, zap.Error(err))
		return err
	}

	return nil
}

// Set
// @description:设置缓存
// @param: key
// @param: data
// @param: expiration
// @return: error
func Set(key string, data interface{}, expiration time.Duration) error {
	// 将对象序列化为字节切片
	serializedData, err := ext.JsonMarshal(data)
	if err != nil {
		global.Logger.Error("redis jsonMarshal error: "+key, zap.Error(err))
		return err
	}

	// 存储字节切片到 Redis
	err = global.Redis.Set(context.Background(), key, serializedData, expiration).Err()
	if err != nil {
		global.Logger.Error("redis set error: "+key, zap.Error(err))
		return err
	}
	return nil
}

// Del
// @description: 删除缓存
// @param: key
// @return: error
func Del(key string) error {
	result, err := global.Redis.Del(context.Background(), key).Result()
	if err != nil {
		global.Logger.Error("redis del error: "+key, zap.Error(err))
		return err
	}
	_ = fmt.Sprintf(strconv.FormatInt(result, 10))
	return nil
}

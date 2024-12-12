package redis

const Nil = RedisError("redis: nil") // nolint:errname

type RedisError string

func (e RedisError) Error() string { return string(e) }

type CacheExpireType int

const (
	/*
		绝对过期
		注：即自创建一段时间后就过期
	*/
	Absolute CacheExpireType = 1
	/*
		相对过期
		注：即该键未被访问后一段时间后过期，若此键一直被访问则过期时间自动延长
	*/
	Relative CacheExpireType = 0
)

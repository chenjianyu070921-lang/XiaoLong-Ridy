package datasource

import (
	"XiaoLong-Ridy/common/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建 go-redis 客户端
func NewRedisClient(c config.RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Host,
		Password: c.Pass,
		DB:       c.Db,
		PoolSize: c.PoolSize,
	})
}

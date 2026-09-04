package datasource

import (
	"context"
	"time"

	"XiaoLong-Ridy/common/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建 go-redis 客户端
func NewRedisClient(c config.RedisConf) *redis.Client {
	dial := time.Duration(c.DialTimeout) * time.Second
	read := time.Duration(c.ReadTimeout) * time.Second
	write := time.Duration(c.WriteTimeout) * time.Second
	if c.DialTimeout <= 0 {
		dial = 5 * time.Second
	}
	if c.ReadTimeout <= 0 {
		read = 3 * time.Second
	}
	if c.WriteTimeout <= 0 {
		write = 3 * time.Second
	}

	return redis.NewClient(&redis.Options{
		Addr:         c.Host,
		Password:     c.Pass,
		DB:           c.Db,
		PoolSize:     c.PoolSize,
		DialTimeout:  dial,
		ReadTimeout:  read,
		WriteTimeout: write,
	})
}

// PingRedis 校验 Redis 连接是否可用
func PingRedis(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}

// MustNewRedisClient 创建客户端并校验连接，连接失败直接 panic（服务启动即暴露问题）
func MustNewRedisClient(c config.RedisConf) *redis.Client {
	client := NewRedisClient(c)
	if err := PingRedis(client); err != nil {
		panic("redis ping failed: " + err.Error())
	}
	return client
}

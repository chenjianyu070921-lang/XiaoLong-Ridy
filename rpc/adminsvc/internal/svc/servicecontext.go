// Package svc 初始化并集中持有 adminsvc 的基础依赖。
package svc

import (
	"database/sql"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

// ServiceContext 是 RPC 逻辑层的依赖容器。
type ServiceContext struct {
	Config config.Config
	MySQL  *sql.DB
	Redis  *redis.Client
}

// NewServiceContext 初始化 MySQL 和 Redis 连接。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := sql.Open("mysql", c.MySQL.DSN)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Cache.Host,
		Password: c.Cache.Password,
		DB:       c.Cache.DB,
	})
	if c.Session.SessionTTLHours <= 0 {
		c.Session.SessionTTLHours = 24
	}
	if c.Session.TokenPrefix == "" {
		c.Session.TokenPrefix = "admin:sess:"
	}
	return &ServiceContext{
		Config: c,
		MySQL:  db,
		Redis:  redisClient,
	}
}

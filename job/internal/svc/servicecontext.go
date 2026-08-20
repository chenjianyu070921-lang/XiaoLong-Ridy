package svc

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/job/internal/config"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	Db             *gorm.DB
	Redis          *redis.Client
	OrderClient    order.Order
	DispatchClient dispatch.Dispatch
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 连接 MySQL
	db, err := datasource.NewMysqlClient(c.Mysql)
	if err != nil {
		panic(fmt.Sprintf("mysql connect failed: %v", err))
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Ping()
	}
	fmt.Println("MySQL 连接成功")

	// 连接 Redis
	redisClient := datasource.NewRedisClient(c.RedisConf)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Sprintf("redis connect failed: %v", err))
	}
	fmt.Println("Redis 连接成功")

	// 创建 ordersvc RPC 客户端（通过 Etcd 服务发现）
	orderClient := order.NewOrder(zrpc.MustNewClient(c.OrderRPC))
	// 创建 dispatchsvc RPC 客户端（通过 Etcd 服务发现）
	dispatchClient := dispatch.NewDispatch(zrpc.MustNewClient(c.DispatchRPC))

	return &ServiceContext{
		Config:         c,
		Db:             db,
		Redis:          redisClient,
		OrderClient:    orderClient,
		DispatchClient: dispatchClient,
	}
}

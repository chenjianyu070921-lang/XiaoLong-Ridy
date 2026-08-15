// Package svc 初始化并集中持有 adminsvc 的基础依赖。
package svc

import (
	"database/sql"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/internal/config"
	driversvcproto "XiaoLong-Ridy/rpc/driversvc/proto"
	ordersvcproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是 RPC 逻辑层共享的依赖容器。
type ServiceContext struct {
	Config config.Config
	MySQL  *sql.DB
	Redis  *redis.Client

	OrdersRPCClient  zrpc.Client
	OrdersSvc        ordersvcproto.OrderClient
	DriversRPCClient zrpc.Client
	DriversSvc       driversvcproto.DriversvcClient
}

// NewServiceContext 初始化 MySQL、Redis 以及下游 RPC 客户端。
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

	ordersClient, ordersSvc, err := newOrdersRPCClient(c.OrdersRPC)
	if err != nil {
		panic(err)
	}
	driversClient, driversSvc, err := newDriversRPCClient(c.DriversRPC)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:           c,
		MySQL:            db,
		Redis:            redisClient,
		OrdersRPCClient:  ordersClient,
		OrdersSvc:        ordersSvc,
		DriversRPCClient: driversClient,
		DriversSvc:       driversSvc,
	}
}

// Close 关闭 RPC 服务依赖。
func (s *ServiceContext) Close() {
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
	if s.MySQL != nil {
		_ = s.MySQL.Close()
	}
}

// newOrdersRPCClient 初始化 ordersvc 客户端。
func newOrdersRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, ordersvcproto.OrderClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:50051"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, ordersvcproto.NewOrderClient(client.Conn()), nil
}

// newDriversRPCClient 初始化 driversvc 客户端。
func newDriversRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, driversvcproto.DriversvcClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:8080"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, driversvcproto.NewDriversvcClient(client.Conn()), nil
}

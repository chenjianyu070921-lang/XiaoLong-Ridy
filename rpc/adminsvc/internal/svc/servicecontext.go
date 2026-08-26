// Package svc 初始化并集中持有 adminsvc 的基础依赖。
package svc

import (
	"database/sql"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/internal/config"
	driversvcproto "XiaoLong-Ridy/rpc/driversvc/proto"
	ordersvcproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	pricesvcproto "XiaoLong-Ridy/rpc/pricesvc/price"
	usersvcproto "XiaoLong-Ridy/rpc/usersvc/proto"

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
	UsersRPCClient   zrpc.Client
	UsersSvc         usersvcproto.UserClient
	DriversRPCClient zrpc.Client
	DriversSvc       driversvcproto.DriversvcClient
	PricesRPCClient  zrpc.Client
	PricesSvc        pricesvcproto.Price
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

	var ordersClient, usersClient, driversClient, pricesClient zrpc.Client
	var ordersSvc ordersvcproto.OrderClient
	var usersSvc usersvcproto.UserClient
	var driversSvc driversvcproto.DriversvcClient
	var pricesSvc pricesvcproto.Price
	// 本地最小服务集无需预建不可达下游连接；默认配置仍完整初始化全部下游 RPC 客户端。
	if !c.DisableDownstreamRPC {
		ordersClient, ordersSvc, err = newOrdersRPCClient(c.OrdersRPC)
		if err != nil {
			panic(err)
		}
		usersClient, usersSvc, err = newUsersRPCClient(c.UsersRPC)
		if err != nil {
			panic(err)
		}
		driversClient, driversSvc, err = newDriversRPCClient(c.DriversRPC)
		if err != nil {
			panic(err)
		}
		pricesClient, pricesSvc, err = newPricesRPCClient(c.PricesRPC)
		if err != nil {
			panic(err)
		}
	}

	return &ServiceContext{
		Config:           c,
		MySQL:            db,
		Redis:            redisClient,
		OrdersRPCClient:  ordersClient,
		OrdersSvc:        ordersSvc,
		UsersRPCClient:   usersClient,
		UsersSvc:         usersSvc,
		DriversRPCClient: driversClient,
		DriversSvc:       driversSvc,
		PricesRPCClient:  pricesClient,
		PricesSvc:        pricesSvc,
	}
}

// newUsersRPCClient 初始化 usersvc 客户端，供后台只读查询用户优惠券历史。
func newUsersRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, usersvcproto.UserClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:50052"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, usersvcproto.NewUserClient(client.Conn()), nil
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
		cfg.Target = "127.0.0.1:5055"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, driversvcproto.NewDriversvcClient(client.Conn()), nil
}

// newPricesRPCClient 初始化 pricesvc 客户端。
func newPricesRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, pricesvcproto.Price, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:50053"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, pricesvcproto.NewPrice(client), nil
}

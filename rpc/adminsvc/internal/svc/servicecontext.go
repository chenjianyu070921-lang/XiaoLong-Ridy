// Package svc 初始化并集中持有 adminsvc 的基础依赖。
package svc

import (
	"database/sql"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/internal/config"
	dispatchsvcproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversvcproto "XiaoLong-Ridy/rpc/driversvc/proto"
	locationsvcproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	ordersvcproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	pricesvcproto "XiaoLong-Ridy/rpc/pricesvc/price"
	pushsvcproto "XiaoLong-Ridy/rpc/pushesvc/pushesvc"
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

	OrdersRPCClient    zrpc.Client
	OrdersSvc          ordersvcproto.OrderClient
	DispatchRPCClient  zrpc.Client
	DispatchSvc        dispatchsvcproto.DispatchClient
	UsersRPCClient     zrpc.Client
	UsersSvc           usersvcproto.UserClient
	DriversRPCClient   zrpc.Client
	DriverSvc          driversvcproto.DriverServiceClient
	LocationsRPCClient zrpc.Client
	LocationSvc        locationsvcproto.LocationServiceClient
	PricesRPCClient    zrpc.Client
	PricesSvc          pricesvcproto.Price
	PushRPCClient      zrpc.Client
	PushSvc            pushsvcproto.PushServiceClient
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

	var ordersClient, dispatchClient, usersClient, driversClient, locationsClient, pricesClient, pushClient zrpc.Client
	var ordersSvc ordersvcproto.OrderClient
	var dispatchSvc dispatchsvcproto.DispatchClient
	var usersSvc usersvcproto.UserClient
	var driversSvc driversvcproto.DriverServiceClient
	var locationsSvc locationsvcproto.LocationServiceClient
	var pricesSvc pricesvcproto.Price
	var pushSvc pushsvcproto.PushServiceClient
	// 本地最小服务集无需预建不可达下游连接；默认配置仍完整初始化全部下游 RPC 客户端。
	if !c.DisableDownstreamRPC {
		ordersClient, ordersSvc, err = newOrdersRPCClient(c.OrdersRPC)
		if err != nil {
			panic(err)
		}
		dispatchClient, dispatchSvc, err = newDispatchRPCClient(c.DispatchRPC)
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
		locationsClient, locationsSvc, err = newLocationsRPCClient(c.LocationsRPC)
		if err != nil {
			panic(err)
		}
		pricesClient, pricesSvc, err = newPricesRPCClient(c.PricesRPC)
		if err != nil {
			panic(err)
		}
		pushClient, pushSvc, err = newPushRPCClient(c.PushRPC)
		if err != nil {
			panic(err)
		}
	}

	return &ServiceContext{
		Config:             c,
		MySQL:              db,
		Redis:              redisClient,
		OrdersRPCClient:    ordersClient,
		OrdersSvc:          ordersSvc,
		DispatchRPCClient:  dispatchClient,
		DispatchSvc:        dispatchSvc,
		UsersRPCClient:     usersClient,
		UsersSvc:           usersSvc,
		DriversRPCClient:   driversClient,
		DriverSvc:          driversSvc,
		LocationsRPCClient: locationsClient,
		LocationSvc:        locationsSvc,
		PricesRPCClient:    pricesClient,
		PricesSvc:          pricesSvc,
		PushRPCClient:      pushClient,
		PushSvc:            pushSvc,
	}
}

// newLocationsRPCClient 初始化 locationsvc 客户端，供后台订单详情轨迹回放查询。
func newLocationsRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, locationsvcproto.LocationServiceClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:9001"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, locationsvcproto.NewLocationServiceClient(client.Conn()), nil
}

// newDispatchRPCClient 初始化 dispatchsvc 客户端，供后台订单详情查询真实派单记录。
func newDispatchRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, dispatchsvcproto.DispatchClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:8083"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, dispatchsvcproto.NewDispatchClient(client.Conn()), nil
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
func newDriversRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, driversvcproto.DriverServiceClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:5055"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, driversvcproto.NewDriverServiceClient(client.Conn()), nil
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

// newPushRPCClient 初始化 pushsvc 客户端，供后台跨服务操作后通知司机端。
func newPushRPCClient(cfg zrpc.RpcClientConf) (zrpc.Client, pushsvcproto.PushServiceClient, error) {
	if len(cfg.Endpoints) == 0 && cfg.Target == "" {
		cfg.Target = "127.0.0.1:9002"
	}
	client, err := zrpc.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, pushsvcproto.NewPushServiceClient(client.Conn()), nil
}

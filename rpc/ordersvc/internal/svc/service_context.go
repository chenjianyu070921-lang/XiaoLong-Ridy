package svc

import (
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config          config.Config
	DB              *gorm.DB
	Redis           *redis.Client
	EventBus        events.Bus
	OrderRepository repository.OrderRepository
	CouponConsumer  repository.CouponConsumer
	DispatchClient  dispatch.Dispatch
	PriceClient     price.Price
	PayClient       pay.Pay
}

// NewServiceContext 初始化 MySQL、Redis、事件总线与订单仓储。
func NewServiceContext(c config.Config) *ServiceContext {
	// mysql 连接
	client, err := datasource.NewMysqlClient(cfg.MysqlConf{
		Dsn:         c.Mysql.DSN,
		MaxOpenConn: 200,
		MaxIdleConn: 30,
		MaxLifeTime: int((time.Hour * 1).Seconds()),
	})
	if err != nil {
		panic(err)
	}

	redisClient := datasource.NewRedisClient(c.Redis)

	dispatchRPC := c.DispatchRPC
	if dispatchRPC.Target == "" && len(dispatchRPC.Endpoints) == 0 {
		dispatchRPC.Target = "127.0.0.1:8083"
	}
	dispatchClient, err := zrpc.NewClient(dispatchRPC)
	if err != nil {
		panic(err)
	}

	priceRPC := c.PriceRPC
	if priceRPC.Target == "" && len(priceRPC.Endpoints) == 0 {
		priceRPC.Target = "127.0.0.1:50053"
	}
	priceClient, err := zrpc.NewClient(priceRPC)
	if err != nil {
		panic(err)
	}

	payRPC := c.PayRPC
	if payRPC.Target == "" && len(payRPC.Endpoints) == 0 {
		payRPC.Target = "127.0.0.1:50054"
	}
	payClient, err := zrpc.NewClient(payRPC)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:          c,
		DB:              client,
		Redis:           redisClient,
		EventBus:        events.NewRedisStreamBus(redisClient, constants.OrderEventStream),
		OrderRepository: repository.NewGormOrderRepository(client),
		CouponConsumer:  repository.NewGormCouponConsumer(client),
		DispatchClient:  dispatch.NewDispatch(dispatchClient),
		PriceClient:     price.NewPrice(priceClient),
		PayClient:       pay.NewPay(payClient),
	}
}

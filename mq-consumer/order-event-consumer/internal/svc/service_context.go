package svc

import (
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/config"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/order"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	Redis          *redis.Client
	EventBus       events.Bus
	DispatchClient dispatch.Dispatch
	OrderClient    order.Order
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := datasource.NewRedisClient(c.Redis)

	dispatchRPC := c.DispatchRPC
	if dispatchRPC.Target == "" && len(dispatchRPC.Endpoints) == 0 {
		dispatchRPC.Target = "127.0.0.1:8083"
	}
	dispatchClient, err := zrpc.NewClient(dispatchRPC)
	if err != nil {
		panic(err)
	}
	orderRPC := c.OrderRPC
	if orderRPC.Target == "" && len(orderRPC.Endpoints) == 0 {
		orderRPC.Target = "127.0.0.1:50051"
	}
	orderClient, err := zrpc.NewClient(orderRPC)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:         c,
		Redis:          redisClient,
		EventBus:       events.NewRedisStreamBus(redisClient, constants.OrderEventStream),
		DispatchClient: dispatch.NewDispatch(dispatchClient),
		OrderClient:    order.NewOrder(orderClient),
	}
}

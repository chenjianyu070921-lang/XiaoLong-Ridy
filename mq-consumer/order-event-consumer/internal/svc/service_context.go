package svc

import (
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/config"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	Redis          *redis.Client
	EventBus       events.Bus
	DispatchClient dispatch.Dispatch
	OrderClient    order.Order
	PayClient      pay.Pay
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := datasource.NewRedisClient(c.Redis)

	dispatchRPC := c.DispatchRPC
	if dispatchRPC.Target == "" && len(dispatchRPC.Endpoints) == 0 {
		dispatchRPC.Target = "127.0.0.1:50056"
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
	payRPC := c.PayRPC
	if payRPC.Target == "" && len(payRPC.Endpoints) == 0 {
		payRPC.Target = "127.0.0.1:50054"
	}
	payClient, err := zrpc.NewClient(payRPC)
	if err != nil {
		// 支付 RPC 不可用时仅降级（退款/支付事件跳过），不影响派单链路。
		logx.Errorf("init pay rpc client failed, pay/refund events disabled: %v", err)
		payClient = nil
	}

	// 事件总线：Kafka ConsumerGroup 消费（与支付模块 paysvc 对齐）。
	// Kafka 未配置或消费者初始化失败时 EventBus 保持 nil，Start 会直接返回错误退出。
	var eventBus events.Bus
	if len(c.Kafka.Brokers) > 0 {
		p, err := mq.NewKafkaProducer(c.Kafka.Brokers)
		if err != nil {
			logx.Errorf("init kafka dlq producer failed: %v", err)
		}
		eventBus = events.NewKafkaBus(nil, mq.NewKafkaConsumer(c.Kafka.Brokers, p).SetOffsetStore(redisClient))
	}

	var payClientSvc pay.Pay
	if payClient != nil {
		payClientSvc = pay.NewPay(payClient)
	}
	return &ServiceContext{
		Config:         c,
		Redis:          redisClient,
		EventBus:       eventBus,
		DispatchClient: dispatch.NewDispatch(dispatchClient),
		OrderClient:    order.NewOrder(orderClient),
		PayClient:      payClientSvc,
	}
}

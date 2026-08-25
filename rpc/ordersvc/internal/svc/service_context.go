package svc

import (
	"XiaoLong-Ridy/rpc/paysvc/pay"
	"time"

	cfg "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/events"
	"XiaoLong-Ridy/common/mq"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                  config.Config
	DB                      *gorm.DB
	Redis                   *redis.Client
	EventBus                events.Bus
	OrderRepository         repository.OrderRepository
	CouponConsumer          repository.CouponConsumer
	RiskBlacklistRepository repository.RiskBlacklistRepository
	DispatchClient          dispatch.Dispatch
	PriceClient             price.Price
	PayClient               pay.Pay
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
	// 支付服务与订单服务存在相互调用场景，使用非阻塞客户端避免本地启动时形成循环依赖。
	// 真正执行支付、退款或结算 RPC 时，gRPC 仍会按请求上下文连接支付服务。
	payRPC.NonBlock = true
	payClient, err := zrpc.NewClient(payRPC, zrpc.WithNonBlock())
	if err != nil {
		panic(err)
	}

	// 事件总线：优先 Kafka（与支付模块 paysvc 对齐），brokers 未配置或生产者初始化失败时
	// EventBus 保持 nil，由调用方回退同步直派（见 create_order_logic 的 published 分支），
	// 避免事件通道不可用时订单创建被阻断。
	var eventBus events.Bus
	if len(c.Kafka.Brokers) > 0 {
		p, err := mq.NewKafkaProducer(c.Kafka.Brokers)
		if err != nil {
			logx.Errorf("init kafka producer failed: %v, fallback to sync dispatch", err)
		} else {
			eventBus = events.NewKafkaBus(p, nil)
		}
	}

	return &ServiceContext{
		Config:                  c,
		DB:                      client,
		Redis:                   redisClient,
		EventBus:                eventBus,
		OrderRepository:         repository.NewGormOrderRepository(client),
		CouponConsumer:          repository.NewGormCouponConsumer(client),
		RiskBlacklistRepository: repository.NewGormRiskBlacklistRepository(client),
		DispatchClient:          dispatch.NewDispatch(dispatchClient),
		PriceClient:             price.NewPrice(priceClient),
		PayClient:               pay.NewPay(payClient),
	}
}

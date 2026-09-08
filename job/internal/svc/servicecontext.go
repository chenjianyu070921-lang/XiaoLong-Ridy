package svc

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/common/datasource"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/job/internal/config"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"
	pushproto "XiaoLong-Ridy/rpc/pushesvc/pushesvc"

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
	PayClient      pay.Pay
	DriverClient   driverproto.DriverServiceClient
	PushClient     pushproto.PushServiceClient
	EventProducer  mq.Producer
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
	// 创建 driversvc RPC 客户端，供管理后台 outbox 冻结司机补偿使用。
	driverRPC := c.DriverRPC
	if len(driverRPC.Endpoints) == 0 && driverRPC.Target == "" {
		driverRPC.Target = "127.0.0.1:50055"
	}
	driverClient := driverproto.NewDriverServiceClient(zrpc.MustNewClient(driverRPC).Conn())
	// 创建 pushsvc RPC 客户端，供管理后台 outbox 通知补偿使用。
	pushRPC := c.PushRPC
	if len(pushRPC.Endpoints) == 0 && pushRPC.Target == "" {
		pushRPC.Target = "127.0.0.1:9002"
	}
	pushClient := pushproto.NewPushServiceClient(zrpc.MustNewClient(pushRPC).Conn())
	// 创建 paysvc RPC 客户端，供支付单创建失败补偿任务使用（P0-3）。
	// 使用非阻塞客户端：job 启动时 paysvc 可能尚未启动，不应阻塞其他定时任务；
	// 真正执行补偿 RPC 时 gRPC 仍会按请求上下文连接，PayClient 为 nil 时任务内部跳过并告警。
	payRPC := c.PayRPC
	if payRPC.Target == "" && len(payRPC.Endpoints) == 0 {
		payRPC.Target = "127.0.0.1:50054"
	}
	payRPC.NonBlock = true
	var payClient pay.Pay
	if payRPCClient, payErr := zrpc.NewClient(payRPC, zrpc.WithNonBlock()); payErr != nil {
		fmt.Printf("paysvc client init failed: %v\n", payErr)
	} else {
		payClient = pay.NewPay(payRPCClient)
	}
	var eventProducer mq.Producer = &mq.NoopProducer{}
	if len(c.Kafka.Brokers) > 0 {
		if producer, producerErr := mq.NewKafkaProducer(c.Kafka.Brokers); producerErr == nil {
			eventProducer = producer
		}
	}

	return &ServiceContext{
		Config:         c,
		Db:             db,
		Redis:          redisClient,
		OrderClient:    orderClient,
		DispatchClient: dispatchClient,
		PayClient:      payClient,
		DriverClient:   driverClient,
		PushClient:     pushClient,
		EventProducer:  eventProducer,
	}
}

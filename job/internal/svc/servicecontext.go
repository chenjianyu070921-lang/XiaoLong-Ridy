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
		driverRPC.Target = "127.0.0.1:5055"
	}
	driverClient := driverproto.NewDriverServiceClient(zrpc.MustNewClient(driverRPC).Conn())
	// 创建 pushsvc RPC 客户端，供管理后台 outbox 通知补偿使用。
	pushRPC := c.PushRPC
	if len(pushRPC.Endpoints) == 0 && pushRPC.Target == "" {
		pushRPC.Target = "127.0.0.1:9002"
	}
	pushClient := pushproto.NewPushServiceClient(zrpc.MustNewClient(pushRPC).Conn())
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
		DriverClient:   driverClient,
		PushClient:     pushClient,
		EventProducer:  eventProducer,
	}
}

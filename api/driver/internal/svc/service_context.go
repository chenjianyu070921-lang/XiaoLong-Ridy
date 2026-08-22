package svc

import (
	"XiaoLong-Ridy/rpc/driversvc/proto"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DriverClient 是 driversvc gRPC 客户端的类型别名，保持与 logic 层命名一致
type DriverClient = proto.DriversvcClient

// ServiceContext 持有下游 gRPC 客户端与 Redis，供 handler/logic 使用
type ServiceContext struct {
	DriverClient   proto.DriversvcClient
	LocationClient locationsvc.LocationServiceClient
	Redis          *redis.Redis
}

// NewServiceContext 构造服务上下文：连接 driversvc / locationsvc gRPC，并初始化 Redis 客户端
func NewServiceContext(driverGRPCAddr, locationGRPCAddr, redisAddr string) *ServiceContext {
	driverConn, err := grpc.Dial(driverGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	locationConn, err := grpc.Dial(locationGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		DriverClient:   proto.NewDriversvcClient(driverConn),
		LocationClient: locationsvc.NewLocationServiceClient(locationConn),
		Redis:          redis.New(redisAddr),
	}
}

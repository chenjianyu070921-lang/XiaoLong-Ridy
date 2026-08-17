// Package svc 定义 driver API 的服务上下文，持有对下游 driversvc 的调用客户端。
package svc

import (
	"context" // 用于在各层之间传递请求上下文与取消信号

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"        // driversvc 的 proto 定义（请求/响应类型）
	"google.golang.org/grpc"                                // gRPC 核心库，用于建立连接
	"google.golang.org/grpc/credentials/insecure"           // 提供非加密（明文）连接凭据，本地联调使用
)

// DriverClient 定义 driver API 调用 driversvc 的公开契约（接口）。
// 当前仅暴露司机「增删改查」四个核心接口，为后续对接留出清晰边界。
type DriverClient interface {
	// CreateDriver 调用创建司机接口。
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	// UpdateDriver 调用更新司机接口。
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	// GetDriver 调用查询司机详情接口。
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	// DeleteDriver 调用删除（软删）司机接口。
	DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error)
}

// grpcClient 是 DriverClient 接口的 gRPC 直连实现，内部持有 driversvc 的 gRPC 客户端。
type grpcClient struct {
	// cli 为 driversvc 生成的 gRPC 客户端实例。
	cli driversproto.DriversvcClient
}

// CreateDriver 转发创建司机请求到 driversvc。
func (g *grpcClient) CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return g.cli.CreateDriver(ctx, req)
}

// UpdateDriver 转发更新司机请求到 driversvc。
func (g *grpcClient) UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return g.cli.UpdateDriver(ctx, req)
}

// GetDriver 转发查询司机请求到 driversvc。
func (g *grpcClient) GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return g.cli.GetDriver(ctx, req)
}

// DeleteDriver 转发删除司机请求到 driversvc。
func (g *grpcClient) DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return g.cli.DeleteDriver(ctx, req)
}

// ServiceContext 持有 driver API 运行所需的依赖（当前为 driversvc 客户端）。
type ServiceContext struct {
	// DriverClient 是与 driversvc 通信的客户端实例，可能为 nil（连接失败时）。
	DriverClient DriverClient
}

// NewServiceContext 构造服务上下文；grpcAddr 为 driversvc 的 gRPC 监听地址。
func NewServiceContext(grpcAddr string) *ServiceContext {
	// 建立到 driversvc 的 gRPC 连接（明文 insecure 凭据，仅本地/内网联调使用）。
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// 连接失败由调用方在首次请求时感知，这里不阻塞启动，仅保留为 nil。
		return &ServiceContext{DriverClient: nil}
	}
	// 连接成功，构造持有 gRPC 客户端的 ServiceContext。
	return &ServiceContext{DriverClient: &grpcClient{cli: driversproto.NewDriversvcClient(conn)}}
}

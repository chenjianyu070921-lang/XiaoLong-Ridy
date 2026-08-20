// Package svc 定义 driver API 的服务上下文，持有对下游 driversvc 的调用客户端。
package svc

import (
	"context" // 用于在各层之间传递请求上下文与取消信号
	"time"    // 验证码缓存有效期

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"        // driversvc 的 proto 定义（请求/响应类型）
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"           // ordersvc 的 proto 定义（接单/行程类型）
	"google.golang.org/grpc"                                // gRPC 核心库，用于建立连接
	"google.golang.org/grpc/credentials/insecure"           // 提供非加密（明文）连接凭据，本地联调使用
)

// DriverClient 定义 driver API 调用 driversvc 的公开契约（接口）。
// 当前暴露司机「增删改查」四个核心接口，以及登录所需的按手机号查询，为后续对接留出清晰边界。
type DriverClient interface {
	// CreateDriver 调用创建司机接口。
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	// UpdateDriver 调用更新司机接口。
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	// GetDriver 调用查询司机详情接口。
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	// GetDriverByPhone 调用按手机号查询司机接口（登录场景）。
	GetDriverByPhone(ctx context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error)
	// SetDriverOnline 调用司机上线接口。
	SetDriverOnline(ctx context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error)
	// SetDriverOffline 调用司机下线接口。
	SetDriverOffline(ctx context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error)
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

// GetDriverByPhone 转发按手机号查询司机请求到 driversvc。
func (g *grpcClient) GetDriverByPhone(ctx context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	return g.cli.GetDriverByPhone(ctx, req)
}

// SetDriverOnline 转发司机上线请求到 driversvc。
func (g *grpcClient) SetDriverOnline(ctx context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	return g.cli.SetDriverOnline(ctx, req)
}

// SetDriverOffline 转发司机下线请求到 driversvc。
func (g *grpcClient) SetDriverOffline(ctx context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	return g.cli.SetDriverOffline(ctx, req)
}

// DeleteDriver 转发删除司机请求到 driversvc。
func (g *grpcClient) DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return g.cli.DeleteDriver(ctx, req)
}

// OrderClient 定义 driver API 调用 ordersvc 的公开契约（接口）。
// 当前暴露司机接单、开始行程、确认到达、结束行程四个接口。
type OrderClient interface {
	// AcceptOrder 调用司机接单接口。
	AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error)
	// StartTrip 调用开始行程接口。
	StartTrip(ctx context.Context, req *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error)
	// ConfirmArrive 调用确认到达接口。
	ConfirmArrive(ctx context.Context, req *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error)
	// FinishTrip 调用结束行程接口。
	FinishTrip(ctx context.Context, req *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error)
}

// orderGRPCClient 是 OrderClient 接口的 gRPC 直连实现。
type orderGRPCClient struct {
	// cli 为 ordersvc 生成的 gRPC 客户端实例。
	cli orderproto.OrderClient
}

// AcceptOrder 转发接单请求到 ordersvc。
func (g *orderGRPCClient) AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	return g.cli.AcceptOrder(ctx, req)
}

// StartTrip 转发开始行程请求到 ordersvc。
func (g *orderGRPCClient) StartTrip(ctx context.Context, req *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error) {
	return g.cli.StartTrip(ctx, req)
}

// ConfirmArrive 转发确认到达请求到 ordersvc。
func (g *orderGRPCClient) ConfirmArrive(ctx context.Context, req *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error) {
	return g.cli.ConfirmArrive(ctx, req)
}

// FinishTrip 转发结束行程请求到 ordersvc。
func (g *orderGRPCClient) FinishTrip(ctx context.Context, req *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error) {
	return g.cli.FinishTrip(ctx, req)
}

// ServiceContext 持有 driver API 运行所需的依赖（当前为 driversvc、ordersvc 客户端）。
type ServiceContext struct {
	// DriverClient 是与 driversvc 通信的客户端实例，可能为 nil（连接失败时）。
	DriverClient DriverClient
	// OrderClient 是与 ordersvc 通信的客户端实例，可能为 nil（连接失败时）。
	OrderClient OrderClient
	// SigningKey 是 JWT 签发/校验的 HMAC-SHA256 密钥，本地联调使用固定值。
	SigningKey string
	// CodeCache 是本地内存验证码缓存，联调阶段顶替短信/缓存服务。
	CodeCache *CodeCache
}

// defaultSigningKey 本地联调阶段的 JWT 签名密钥，与 passenger 端保持一致以便统一调试。
const defaultSigningKey = "local-development-signing-key"

// defaultCodeTTL 验证码默认有效期（5 分钟）。
const defaultCodeTTL = 5 * time.Minute

// NewServiceContext 构造服务上下文；driverGRPCAddr、orderGRPCAddr 分别为下游服务地址。
func NewServiceContext(driverGRPCAddr, orderGRPCAddr string) *ServiceContext {
	// 建立到 driversvc 的 gRPC 连接（明文 insecure 凭据，仅本地/内网联调使用）。
	driverConn, driverErr := grpc.NewClient(driverGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// 建立到 ordersvc 的 gRPC 连接。
	orderConn, orderErr := grpc.NewClient(orderGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 两个下游均未连上时不阻塞启动，客户端置空由调用方在首次请求感知。
	if driverErr != nil && orderErr != nil {
		return &ServiceContext{
			DriverClient: nil,
			OrderClient:  nil,
			SigningKey:   defaultSigningKey,
			CodeCache:    NewCodeCache(defaultCodeTTL),
		}
	}

	svcCtx := &ServiceContext{
		SigningKey: defaultSigningKey,
		CodeCache:  NewCodeCache(defaultCodeTTL),
	}
	// driversvc 连接成功则注入客户端，失败保留 nil。
	if driverErr == nil {
		svcCtx.DriverClient = &grpcClient{cli: driversproto.NewDriversvcClient(driverConn)}
	}
	// ordersvc 连接成功则注入客户端，失败保留 nil。
	if orderErr == nil {
		svcCtx.OrderClient = &orderGRPCClient{cli: orderproto.NewOrderClient(orderConn)}
	}
	return svcCtx
}

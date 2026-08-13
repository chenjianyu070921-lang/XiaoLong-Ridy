// Package svc 定义 driver API 的服务上下文，持有对下游 driversvc 的调用客户端。
package svc

import (
	"context" // 用于在各层之间传递请求上下文与取消信号

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"        // driversvc 的 proto 定义（请求/响应类型）
	"google.golang.org/grpc"                                // gRPC 核心库，用于建立连接
	"google.golang.org/grpc/credentials/insecure"           // 提供非加密（明文）连接凭据，本地联调使用
)

// DriverClient 定义 driver API 调用 driversvc 的公开契约（接口）。
// 使用接口便于后续替换为 mock 实现或切换底层通信方式。
type DriverClient interface {
	// Ping 调用 driversvc 的健康检查接口。
	Ping(ctx context.Context, req *driversproto.Request) (*driversproto.Response, error)
	// CreateDriver 调用创建司机接口。
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	// UpdateDriver 调用更新司机接口。
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	// DeleteDriver 调用删除（软删）司机接口。
	DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error)
	// GetDriver 调用查询司机详情接口。
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	// ListDrivers 调用分页查询司机列表接口。
	ListDrivers(ctx context.Context, req *driversproto.ListDriversRequest) (*driversproto.ListDriversResponse, error)
	// CreateVehicle 调用创建车辆接口。
	CreateVehicle(ctx context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error)
	// UpdateVehicle 调用更新车辆接口。
	UpdateVehicle(ctx context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error)
	// DeleteVehicle 调用删除车辆接口。
	DeleteVehicle(ctx context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error)
	// GetVehicle 调用查询车辆详情接口。
	GetVehicle(ctx context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error)
	// ListVehicles 调用分页查询车辆列表接口。
	ListVehicles(ctx context.Context, req *driversproto.ListVehiclesRequest) (*driversproto.ListVehiclesResponse, error)
	// CreateCertification 调用创建认证接口。
	CreateCertification(ctx context.Context, req *driversproto.CreateCertificationRequest) (*driversproto.CreateCertificationResponse, error)
	// UpdateCertification 调用更新认证接口（含审核状态流转）。
	UpdateCertification(ctx context.Context, req *driversproto.UpdateCertificationRequest) (*driversproto.UpdateCertificationResponse, error)
	// DeleteCertification 调用删除认证接口。
	DeleteCertification(ctx context.Context, req *driversproto.DeleteCertificationRequest) (*driversproto.DeleteCertificationResponse, error)
	// GetCertification 调用查询认证详情接口。
	GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error)
	// ListCertifications 调用分页查询认证列表接口。
	ListCertifications(ctx context.Context, req *driversproto.ListCertificationsRequest) (*driversproto.ListCertificationsResponse, error)
	// CreateScore 调用创建服务分接口。
	CreateScore(ctx context.Context, req *driversproto.CreateScoreRequest) (*driversproto.CreateScoreResponse, error)
	// UpdateScore 调用更新服务分接口。
	UpdateScore(ctx context.Context, req *driversproto.UpdateScoreRequest) (*driversproto.UpdateScoreResponse, error)
	// DeleteScore 调用删除服务分接口。
	DeleteScore(ctx context.Context, req *driversproto.DeleteScoreRequest) (*driversproto.DeleteScoreResponse, error)
	// GetScore 调用查询服务分详情接口。
	GetScore(ctx context.Context, req *driversproto.GetScoreRequest) (*driversproto.GetScoreResponse, error)
	// ListScores 调用分页查询服务分列表接口。
	ListScores(ctx context.Context, req *driversproto.ListScoresRequest) (*driversproto.ListScoresResponse, error)
	// CreateWithdraw 调用创建提现接口。
	CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error)
	// UpdateWithdraw 调用更新提现接口（含打款状态流转）。
	UpdateWithdraw(ctx context.Context, req *driversproto.UpdateWithdrawRequest) (*driversproto.UpdateWithdrawResponse, error)
	// DeleteWithdraw 调用删除提现接口。
	DeleteWithdraw(ctx context.Context, req *driversproto.DeleteWithdrawRequest) (*driversproto.DeleteWithdrawResponse, error)
	// GetWithdraw 调用查询提现详情接口。
	GetWithdraw(ctx context.Context, req *driversproto.GetWithdrawRequest) (*driversproto.GetWithdrawResponse, error)
	// ListWithdraws 调用分页查询提现列表接口。
	ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error)
}

// grpcClient 是 DriverClient 接口的 gRPC 直连实现，内部持有 driversvc 的 gRPC 客户端。
type grpcClient struct {
	// cli 为 driversvc 生成的 gRPC 客户端实例。
	cli driversproto.DriversvcClient
}

// Ping 转发健康检查请求到 driversvc。
func (g *grpcClient) Ping(ctx context.Context, req *driversproto.Request) (*driversproto.Response, error) {
	return g.cli.Ping(ctx, req)
}

// CreateDriver 转发创建司机请求到 driversvc。
func (g *grpcClient) CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return g.cli.CreateDriver(ctx, req)
}

// UpdateDriver 转发更新司机请求到 driversvc。
func (g *grpcClient) UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return g.cli.UpdateDriver(ctx, req)
}

// DeleteDriver 转发删除司机请求到 driversvc。
func (g *grpcClient) DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return g.cli.DeleteDriver(ctx, req)
}

// GetDriver 转发查询司机请求到 driversvc。
func (g *grpcClient) GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return g.cli.GetDriver(ctx, req)
}

// ListDrivers 转发分页查询司机请求到 driversvc。
func (g *grpcClient) ListDrivers(ctx context.Context, req *driversproto.ListDriversRequest) (*driversproto.ListDriversResponse, error) {
	return g.cli.ListDrivers(ctx, req)
}

// CreateVehicle 转发创建车辆请求到 driversvc。
func (g *grpcClient) CreateVehicle(ctx context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	return g.cli.CreateVehicle(ctx, req)
}

// UpdateVehicle 转发更新车辆请求到 driversvc。
func (g *grpcClient) UpdateVehicle(ctx context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	return g.cli.UpdateVehicle(ctx, req)
}

// DeleteVehicle 转发删除车辆请求到 driversvc。
func (g *grpcClient) DeleteVehicle(ctx context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	return g.cli.DeleteVehicle(ctx, req)
}

// GetVehicle 转发查询车辆请求到 driversvc。
func (g *grpcClient) GetVehicle(ctx context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	return g.cli.GetVehicle(ctx, req)
}

// ListVehicles 转发分页查询车辆请求到 driversvc。
func (g *grpcClient) ListVehicles(ctx context.Context, req *driversproto.ListVehiclesRequest) (*driversproto.ListVehiclesResponse, error) {
	return g.cli.ListVehicles(ctx, req)
}

// CreateCertification 转发创建认证请求到 driversvc。
func (g *grpcClient) CreateCertification(ctx context.Context, req *driversproto.CreateCertificationRequest) (*driversproto.CreateCertificationResponse, error) {
	return g.cli.CreateCertification(ctx, req)
}

// UpdateCertification 转发更新认证请求到 driversvc。
func (g *grpcClient) UpdateCertification(ctx context.Context, req *driversproto.UpdateCertificationRequest) (*driversproto.UpdateCertificationResponse, error) {
	return g.cli.UpdateCertification(ctx, req)
}

// DeleteCertification 转发删除认证请求到 driversvc。
func (g *grpcClient) DeleteCertification(ctx context.Context, req *driversproto.DeleteCertificationRequest) (*driversproto.DeleteCertificationResponse, error) {
	return g.cli.DeleteCertification(ctx, req)
}

// GetCertification 转发查询认证请求到 driversvc。
func (g *grpcClient) GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return g.cli.GetCertification(ctx, req)
}

// ListCertifications 转发分页查询认证请求到 driversvc。
func (g *grpcClient) ListCertifications(ctx context.Context, req *driversproto.ListCertificationsRequest) (*driversproto.ListCertificationsResponse, error) {
	return g.cli.ListCertifications(ctx, req)
}

// CreateScore 转发创建服务分请求到 driversvc。
func (g *grpcClient) CreateScore(ctx context.Context, req *driversproto.CreateScoreRequest) (*driversproto.CreateScoreResponse, error) {
	return g.cli.CreateScore(ctx, req)
}

// UpdateScore 转发更新服务分请求到 driversvc。
func (g *grpcClient) UpdateScore(ctx context.Context, req *driversproto.UpdateScoreRequest) (*driversproto.UpdateScoreResponse, error) {
	return g.cli.UpdateScore(ctx, req)
}

// DeleteScore 转发删除服务分请求到 driversvc。
func (g *grpcClient) DeleteScore(ctx context.Context, req *driversproto.DeleteScoreRequest) (*driversproto.DeleteScoreResponse, error) {
	return g.cli.DeleteScore(ctx, req)
}

// GetScore 转发查询服务分请求到 driversvc。
func (g *grpcClient) GetScore(ctx context.Context, req *driversproto.GetScoreRequest) (*driversproto.GetScoreResponse, error) {
	return g.cli.GetScore(ctx, req)
}

// ListScores 转发分页查询服务分请求到 driversvc。
func (g *grpcClient) ListScores(ctx context.Context, req *driversproto.ListScoresRequest) (*driversproto.ListScoresResponse, error) {
	return g.cli.ListScores(ctx, req)
}

// CreateWithdraw 转发创建提现请求到 driversvc。
func (g *grpcClient) CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return g.cli.CreateWithdraw(ctx, req)
}

// UpdateWithdraw 转发更新提现请求到 driversvc。
func (g *grpcClient) UpdateWithdraw(ctx context.Context, req *driversproto.UpdateWithdrawRequest) (*driversproto.UpdateWithdrawResponse, error) {
	return g.cli.UpdateWithdraw(ctx, req)
}

// DeleteWithdraw 转发删除提现请求到 driversvc。
func (g *grpcClient) DeleteWithdraw(ctx context.Context, req *driversproto.DeleteWithdrawRequest) (*driversproto.DeleteWithdrawResponse, error) {
	return g.cli.DeleteWithdraw(ctx, req)
}

// GetWithdraw 转发查询提现请求到 driversvc。
func (g *grpcClient) GetWithdraw(ctx context.Context, req *driversproto.GetWithdrawRequest) (*driversproto.GetWithdrawResponse, error) {
	return g.cli.GetWithdraw(ctx, req)
}

// ListWithdraws 转发分页查询提现请求到 driversvc。
func (g *grpcClient) ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return g.cli.ListWithdraws(ctx, req)
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

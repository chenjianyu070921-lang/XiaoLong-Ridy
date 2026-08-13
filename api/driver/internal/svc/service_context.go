package svc

import (
	"context"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DriverClient 定义 driver API 调用 driversvc 的公开契约。
type DriverClient interface {
	Ping(ctx context.Context, req *driversproto.Request) (*driversproto.Response, error)
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error)
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	ListDrivers(ctx context.Context, req *driversproto.ListDriversRequest) (*driversproto.ListDriversResponse, error)
	CreateVehicle(ctx context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error)
	UpdateVehicle(ctx context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error)
	DeleteVehicle(ctx context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error)
	GetVehicle(ctx context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error)
	ListVehicles(ctx context.Context, req *driversproto.ListVehiclesRequest) (*driversproto.ListVehiclesResponse, error)
	CreateCertification(ctx context.Context, req *driversproto.CreateCertificationRequest) (*driversproto.CreateCertificationResponse, error)
	UpdateCertification(ctx context.Context, req *driversproto.UpdateCertificationRequest) (*driversproto.UpdateCertificationResponse, error)
	DeleteCertification(ctx context.Context, req *driversproto.DeleteCertificationRequest) (*driversproto.DeleteCertificationResponse, error)
	GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error)
	ListCertifications(ctx context.Context, req *driversproto.ListCertificationsRequest) (*driversproto.ListCertificationsResponse, error)
	CreateScore(ctx context.Context, req *driversproto.CreateScoreRequest) (*driversproto.CreateScoreResponse, error)
	UpdateScore(ctx context.Context, req *driversproto.UpdateScoreRequest) (*driversproto.UpdateScoreResponse, error)
	DeleteScore(ctx context.Context, req *driversproto.DeleteScoreRequest) (*driversproto.DeleteScoreResponse, error)
	GetScore(ctx context.Context, req *driversproto.GetScoreRequest) (*driversproto.GetScoreResponse, error)
	ListScores(ctx context.Context, req *driversproto.ListScoresRequest) (*driversproto.ListScoresResponse, error)
	CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error)
	UpdateWithdraw(ctx context.Context, req *driversproto.UpdateWithdrawRequest) (*driversproto.UpdateWithdrawResponse, error)
	DeleteWithdraw(ctx context.Context, req *driversproto.DeleteWithdrawRequest) (*driversproto.DeleteWithdrawResponse, error)
	GetWithdraw(ctx context.Context, req *driversproto.GetWithdrawRequest) (*driversproto.GetWithdrawResponse, error)
	ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error)
}

// grpcClient 是 DriverClient 的 gRPC 直连实现。
type grpcClient struct {
	cli driversproto.DriversvcClient
}

func (g *grpcClient) Ping(ctx context.Context, req *driversproto.Request) (*driversproto.Response, error) {
	return g.cli.Ping(ctx, req)
}
func (g *grpcClient) CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return g.cli.CreateDriver(ctx, req)
}
func (g *grpcClient) UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return g.cli.UpdateDriver(ctx, req)
}
func (g *grpcClient) DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return g.cli.DeleteDriver(ctx, req)
}
func (g *grpcClient) GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return g.cli.GetDriver(ctx, req)
}
func (g *grpcClient) ListDrivers(ctx context.Context, req *driversproto.ListDriversRequest) (*driversproto.ListDriversResponse, error) {
	return g.cli.ListDrivers(ctx, req)
}
func (g *grpcClient) CreateVehicle(ctx context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	return g.cli.CreateVehicle(ctx, req)
}
func (g *grpcClient) UpdateVehicle(ctx context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	return g.cli.UpdateVehicle(ctx, req)
}
func (g *grpcClient) DeleteVehicle(ctx context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	return g.cli.DeleteVehicle(ctx, req)
}
func (g *grpcClient) GetVehicle(ctx context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	return g.cli.GetVehicle(ctx, req)
}
func (g *grpcClient) ListVehicles(ctx context.Context, req *driversproto.ListVehiclesRequest) (*driversproto.ListVehiclesResponse, error) {
	return g.cli.ListVehicles(ctx, req)
}
func (g *grpcClient) CreateCertification(ctx context.Context, req *driversproto.CreateCertificationRequest) (*driversproto.CreateCertificationResponse, error) {
	return g.cli.CreateCertification(ctx, req)
}
func (g *grpcClient) UpdateCertification(ctx context.Context, req *driversproto.UpdateCertificationRequest) (*driversproto.UpdateCertificationResponse, error) {
	return g.cli.UpdateCertification(ctx, req)
}
func (g *grpcClient) DeleteCertification(ctx context.Context, req *driversproto.DeleteCertificationRequest) (*driversproto.DeleteCertificationResponse, error) {
	return g.cli.DeleteCertification(ctx, req)
}
func (g *grpcClient) GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return g.cli.GetCertification(ctx, req)
}
func (g *grpcClient) ListCertifications(ctx context.Context, req *driversproto.ListCertificationsRequest) (*driversproto.ListCertificationsResponse, error) {
	return g.cli.ListCertifications(ctx, req)
}
func (g *grpcClient) CreateScore(ctx context.Context, req *driversproto.CreateScoreRequest) (*driversproto.CreateScoreResponse, error) {
	return g.cli.CreateScore(ctx, req)
}
func (g *grpcClient) UpdateScore(ctx context.Context, req *driversproto.UpdateScoreRequest) (*driversproto.UpdateScoreResponse, error) {
	return g.cli.UpdateScore(ctx, req)
}
func (g *grpcClient) DeleteScore(ctx context.Context, req *driversproto.DeleteScoreRequest) (*driversproto.DeleteScoreResponse, error) {
	return g.cli.DeleteScore(ctx, req)
}
func (g *grpcClient) GetScore(ctx context.Context, req *driversproto.GetScoreRequest) (*driversproto.GetScoreResponse, error) {
	return g.cli.GetScore(ctx, req)
}
func (g *grpcClient) ListScores(ctx context.Context, req *driversproto.ListScoresRequest) (*driversproto.ListScoresResponse, error) {
	return g.cli.ListScores(ctx, req)
}
func (g *grpcClient) CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return g.cli.CreateWithdraw(ctx, req)
}
func (g *grpcClient) UpdateWithdraw(ctx context.Context, req *driversproto.UpdateWithdrawRequest) (*driversproto.UpdateWithdrawResponse, error) {
	return g.cli.UpdateWithdraw(ctx, req)
}
func (g *grpcClient) DeleteWithdraw(ctx context.Context, req *driversproto.DeleteWithdrawRequest) (*driversproto.DeleteWithdrawResponse, error) {
	return g.cli.DeleteWithdraw(ctx, req)
}
func (g *grpcClient) GetWithdraw(ctx context.Context, req *driversproto.GetWithdrawRequest) (*driversproto.GetWithdrawResponse, error) {
	return g.cli.GetWithdraw(ctx, req)
}
func (g *grpcClient) ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return g.cli.ListWithdraws(ctx, req)
}

type ServiceContext struct {
	DriverClient DriverClient
}

// NewServiceContext 构造服务上下文；grpcAddr 为 driversvc 的 gRPC 监听地址。
func NewServiceContext(grpcAddr string) *ServiceContext {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// 连接失败由调用方在首次请求时感知，这里不阻塞启动。
		return &ServiceContext{DriverClient: nil}
	}
	return &ServiceContext{DriverClient: &grpcClient{cli: driversproto.NewDriversvcClient(conn)}}
}

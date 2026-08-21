package svc

import (
	"context"
	"time"

	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverClient interface {
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	RegisterDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	GetDriverByPhone(ctx context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error)
	SetDriverOnline(ctx context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error)
	SetDriverOffline(ctx context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error)
	ReportLocation(ctx context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error)
	SetDriverServiceStatus(ctx context.Context, req *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error)
	Heartbeat(ctx context.Context, req *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error)
	DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error)
	Login(ctx context.Context, req *driversproto.LoginRequest) (*driversproto.LoginResponse, error)
	LoginBySMS(ctx context.Context, req *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error)
	GetDriverAiScore(ctx context.Context, req *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error)
	UploadCertification(ctx context.Context, req *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error)
	GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error)
}

type grpcClient struct {
	cli driversproto.DriversvcClient
}

func (g *grpcClient) CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return g.cli.CreateDriver(ctx, req)
}

func (g *grpcClient) RegisterDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return g.cli.RegisterDriver(ctx, req)
}

func (g *grpcClient) UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return g.cli.UpdateDriver(ctx, req)
}

func (g *grpcClient) GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return g.cli.GetDriver(ctx, req)
}

func (g *grpcClient) GetDriverByPhone(ctx context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	return g.cli.GetDriverByPhone(ctx, req)
}

func (g *grpcClient) SetDriverOnline(ctx context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	return g.cli.SetDriverOnline(ctx, req)
}

func (g *grpcClient) SetDriverOffline(ctx context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	return g.cli.SetDriverOffline(ctx, req)
}

func (g *grpcClient) ReportLocation(ctx context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	return g.cli.ReportLocation(ctx, req)
}

func (g *grpcClient) SetDriverServiceStatus(ctx context.Context, req *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error) {
	return g.cli.SetDriverServiceStatus(ctx, req)
}

func (g *grpcClient) Heartbeat(ctx context.Context, req *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error) {
	return g.cli.Heartbeat(ctx, req)
}

func (g *grpcClient) DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return g.cli.DeleteDriver(ctx, req)
}

func (g *grpcClient) Login(ctx context.Context, req *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	return g.cli.Login(ctx, req)
}

func (g *grpcClient) LoginBySMS(ctx context.Context, req *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	return g.cli.LoginBySMS(ctx, req)
}

func (g *grpcClient) GetDriverAiScore(ctx context.Context, req *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error) {
	return g.cli.GetDriverAiScore(ctx, req)
}

func (g *grpcClient) UploadCertification(ctx context.Context, req *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error) {
	return g.cli.UploadCertification(ctx, req)
}

func (g *grpcClient) GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return g.cli.GetCertification(ctx, req)
}

type OrderClient interface {
	GetOrder(ctx context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error)
	AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error)
	StartTrip(ctx context.Context, req *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error)
	ConfirmArrive(ctx context.Context, req *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error)
	FinishTrip(ctx context.Context, req *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error)
}

type orderGRPCClient struct {
	cli orderproto.OrderClient
}

func (g *orderGRPCClient) GetOrder(ctx context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	return g.cli.GetOrder(ctx, req)
}

func (g *orderGRPCClient) AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	return g.cli.AcceptOrder(ctx, req)
}

func (g *orderGRPCClient) StartTrip(ctx context.Context, req *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error) {
	return g.cli.StartTrip(ctx, req)
}

func (g *orderGRPCClient) ConfirmArrive(ctx context.Context, req *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error) {
	return g.cli.ConfirmArrive(ctx, req)
}

func (g *orderGRPCClient) FinishTrip(ctx context.Context, req *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error) {
	return g.cli.FinishTrip(ctx, req)
}

type DispatchClient interface {
	RejectDispatch(ctx context.Context, req *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error)
	ListDispatchRecords(ctx context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error)
}

type dispatchGRPCClient struct {
	cli dispatchproto.DispatchClient
}

func (g *dispatchGRPCClient) RejectDispatch(ctx context.Context, req *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error) {
	return g.cli.RejectDispatch(ctx, req)
}

func (g *dispatchGRPCClient) ListDispatchRecords(ctx context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	return g.cli.ListDispatchRecords(ctx, req)
}

type ServiceContext struct {
	DriverClient   DriverClient
	OrderClient    OrderClient
	DispatchClient DispatchClient
	SigningKey     string
	CodeCache      *CodeCache
}

const defaultSigningKey = "local-development-signing-key"

const defaultCodeTTL = 5 * time.Minute

func NewServiceContext(driverGRPCAddr, orderGRPCAddr string) *ServiceContext {
	driverConn, driverErr := grpc.NewClient(driverGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	orderConn, orderErr := grpc.NewClient(orderGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	svcCtx := &ServiceContext{
		SigningKey: defaultSigningKey,
		CodeCache:  NewCodeCache(defaultCodeTTL),
	}
	if driverErr == nil {
		svcCtx.DriverClient = &grpcClient{cli: driversproto.NewDriversvcClient(driverConn)}
	}
	if orderErr == nil {
		svcCtx.OrderClient = &orderGRPCClient{cli: orderproto.NewOrderClient(orderConn)}
	}
	return svcCtx
}

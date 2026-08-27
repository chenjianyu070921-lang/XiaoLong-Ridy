package svc

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	commonconfig "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// defaultRPCTimeout 是 gRPC 调用的默认超时，下游不可用时快速失败而非 hang。
const defaultRPCTimeout = 3 * time.Second

// timeoutInterceptor 为未设置 deadline 的 gRPC 调用统一加超时。
func timeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// grpcDialOpts 是所有 gRPC 客户端共用的拨号选项：不安全凭证 + 统一超时拦截器。
var grpcDialOpts = []grpc.DialOption{
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithChainUnaryInterceptor(timeoutInterceptor(defaultRPCTimeout)),
}

type DriverClient interface {
	CreateDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	RegisterDriver(ctx context.Context, req *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error)
	UpdateDriver(ctx context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error)
	GetDriver(ctx context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error)
	GetDriverByPhone(ctx context.Context, req *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error)
	SetDriverOnline(ctx context.Context, req *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error)
	SetDriverOffline(ctx context.Context, req *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error)
	SetDriverListenPreference(ctx context.Context, req *driversproto.SetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error)
	GetDriverListenPreference(ctx context.Context, req *driversproto.GetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error)
	ReportLocation(ctx context.Context, req *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error)
	SetDriverServiceStatus(ctx context.Context, req *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error)
	Heartbeat(ctx context.Context, req *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error)
	DeleteDriver(ctx context.Context, req *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error)
	Login(ctx context.Context, req *driversproto.LoginRequest) (*driversproto.LoginResponse, error)
	LoginBySMS(ctx context.Context, req *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error)
	CreateVehicle(ctx context.Context, req *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error)
	UpdateVehicle(ctx context.Context, req *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error)
	DeleteVehicle(ctx context.Context, req *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error)
	GetVehicle(ctx context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error)
	ListNearbyDrivers(ctx context.Context, req *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error)
	GetDriverAiScore(ctx context.Context, req *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error)
	UploadCertification(ctx context.Context, req *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error)
	GetCertification(ctx context.Context, req *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error)
	CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error)
	ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error)
}

type grpcClient struct {
	cli driversproto.DriverServiceClient
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

func (g *grpcClient) SetDriverListenPreference(ctx context.Context, req *driversproto.SetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error) {
	return g.cli.SetDriverListenPreference(ctx, req)
}

func (g *grpcClient) GetDriverListenPreference(ctx context.Context, req *driversproto.GetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error) {
	return g.cli.GetDriverListenPreference(ctx, req)
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
	return g.cli.LoginBySms(ctx, req)
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

func (g *grpcClient) ListNearbyDrivers(ctx context.Context, req *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error) {
	return g.cli.ListNearbyDrivers(ctx, req)
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

func (g *grpcClient) CreateWithdraw(ctx context.Context, req *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return g.cli.CreateWithdraw(ctx, req)
}

func (g *grpcClient) ListWithdraws(ctx context.Context, req *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return g.cli.ListWithdraws(ctx, req)
}

type OrderClient interface {
	GetOrder(ctx context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error)
	ListOrders(ctx context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error)
	AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error)
	CancelOrder(ctx context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error)
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

func (g *orderGRPCClient) ListOrders(ctx context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	return g.cli.ListOrders(ctx, req)
}

func (g *orderGRPCClient) AcceptOrder(ctx context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	return g.cli.AcceptOrder(ctx, req)
}

func (g *orderGRPCClient) CancelOrder(ctx context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	return g.cli.CancelOrder(ctx, req)
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

type LocationClient interface {
	ReportLocation(ctx context.Context, req *locationproto.ReportLocationReq) (*locationproto.ReportLocationResp, error)
}

type locationGRPCClient struct {
	cli locationproto.LocationServiceClient
}

func (g *locationGRPCClient) ReportLocation(ctx context.Context, req *locationproto.ReportLocationReq) (*locationproto.ReportLocationResp, error) {
	return g.cli.ReportLocation(ctx, req)
}

type ServiceContext struct {
	DriverClient         DriverClient
	OrderClient          OrderClient
	DispatchClient       DispatchClient
	LocationClient       LocationClient
	ReviewRepository     ReviewRepository
	TrajectoryRepository TrajectoryRepository
	SigningKey           string
	CodeCache            CodeCache
	RedisClient          *redis.Client
	PushPollInterval     time.Duration
	PushPollPageSize     int32
}

const defaultSigningKey = "local-development-signing-key"

const defaultCodeTTL = 5 * time.Minute

func NewServiceContext(driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr, locationGRPCAddr, redisAddr string) *ServiceContext {
	return NewServiceContextWithStorage(driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr, locationGRPCAddr, redisAddr, commonconfig.MysqlConf{})
}

func NewServiceContextWithStorage(driverGRPCAddr, orderGRPCAddr, dispatchGRPCAddr, locationGRPCAddr, redisAddr string, mysqlConf commonconfig.MysqlConf) *ServiceContext {
	driverConn, driverErr := grpc.NewClient(driverGRPCAddr, grpcDialOpts...)
	orderConn, orderErr := grpc.NewClient(orderGRPCAddr, grpcDialOpts...)
	dispatchConn, dispatchErr := grpc.NewClient(dispatchGRPCAddr, grpcDialOpts...)
	locationConn, locationErr := grpc.NewClient(locationGRPCAddr, grpcDialOpts...)

	// Code cache: use Redis when configured, otherwise fall back to local memory.
	var codeCache CodeCache
	var rdb *redis.Client
	if redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
		codeCache = NewRedisCodeCache(rdb, defaultCodeTTL)
	} else {
		codeCache = NewLocalCodeCache(defaultCodeTTL)
	}

	svcCtx := &ServiceContext{
		SigningKey:  resolveSigningKey(),
		CodeCache:   codeCache,
		RedisClient: rdb,
	}
	if driverErr == nil {
		svcCtx.DriverClient = &grpcClient{cli: driversproto.NewDriverServiceClient(driverConn)}
	}
	if orderErr == nil {
		svcCtx.OrderClient = &orderGRPCClient{cli: orderproto.NewOrderClient(orderConn)}
	}
	if dispatchErr == nil {
		svcCtx.DispatchClient = &dispatchGRPCClient{cli: dispatchproto.NewDispatchClient(dispatchConn)}
	}
	if locationErr == nil {
		svcCtx.LocationClient = &locationGRPCClient{cli: locationproto.NewLocationServiceClient(locationConn)}
	}
	if strings.TrimSpace(mysqlConf.Dsn) != "" {
		db, err := datasource.NewMysqlClient(mysqlConf)
		if err != nil {
			// 妥协：MySQL 初始化失败不 panic，打日志后继续启动，Review/Trajectory 接口返回 501 降级。
			logx.Errorf("driver api mysql init failed, review/trajectory endpoints will be unavailable: %v", err)
		} else {
			svcCtx.ReviewRepository = NewGormReviewRepository(db)
			svcCtx.TrajectoryRepository = NewGormTrajectoryRepository(db)
		}
	}
	return svcCtx
}

func resolveSigningKey() string {
	if key := strings.TrimSpace(os.Getenv("DRIVER_SIGNING_KEY")); key != "" {
		return key
	}
	return defaultSigningKey
}

func (s *ServiceContext) ValidateSigningKey() error {
	if s == nil || strings.TrimSpace(s.SigningKey) == "" {
		return errors.New("driver signing key is empty")
	}
	if expected := strings.TrimSpace(os.Getenv("DRIVERSVC_SIGNING_KEY")); expected != "" && expected != s.SigningKey {
		return errors.New("DRIVER_SIGNING_KEY and DRIVERSVC_SIGNING_KEY mismatch")
	}
	return nil
}

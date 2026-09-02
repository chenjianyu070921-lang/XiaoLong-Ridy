package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	commonconfig "XiaoLong-Ridy/common/config"
	"XiaoLong-Ridy/common/datasource"
	qiniuutil "XiaoLong-Ridy/common/qiniu"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	orderlocal "XiaoLong-Ridy/rpc/ordersvc/client"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	paylocal "XiaoLong-Ridy/rpc/paysvc/pay"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	userlocal "XiaoLong-Ridy/rpc/usersvc/client"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
)

const (
	defaultHTTPAddr        = ":8091"
	clientModeGRPC         = "grpc"
	clientModeLocal        = "local"
	defaultUserRPCAddr     = "127.0.0.1:50052"
	defaultOrderRPCAddr    = "127.0.0.1:50051"
	defaultPriceRPCAddr    = "127.0.0.1:50053"
	defaultPayRPCAddr      = "127.0.0.1:50054"
	defaultDispatchRPCAddr = "127.0.0.1:50056"
	defaultLocationRPCAddr = "127.0.0.1:9001"
	defaultPriceCityCode   = "110000"
	localDevSigningKey     = "xiaolong-passenger-local-dev-key"
)

// RuntimeConfig 保存 passenger API 启动时需要的运行参数。
// RPC 地址为空时会补充本地默认 gRPC 地址，运行时不再创建本地内存客户端。
type RuntimeConfig struct {
	HTTPAddr        string
	TokenSigningKey string
	UserRPCAddr     string
	OrderRPCAddr    string
	PriceRPCAddr    string
	PayRPCAddr      string
	DispatchRPCAddr string
	LocationRPCAddr string
	ClientMode      string
	PriceCityCode   string
	MysqlDSN        string
	QiniuAccessKey  string
	QiniuSecretKey  string
	QiniuBucket     string
	QiniuDomain     string
	QiniuUploadURL  string
}

// UserClient 定义 passenger API 调用 usersvc 的完整 RPC 契约。
type UserClient interface {
	SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error)
	LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error)
	LoginByPassword(ctx context.Context, req *userproto.LoginByPasswordRequest) (*userproto.LoginBySMSResponse, error)
	SetPassword(ctx context.Context, req *userproto.SetPasswordRequest) (*userproto.SetPasswordResponse, error)
	RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error)
	Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error)
	GetProfile(ctx context.Context, req *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error)
	SubmitRealName(ctx context.Context, req *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error)
	UpdateProfile(ctx context.Context, req *userproto.UpdateProfileRequest) (*userproto.UpdateProfileResponse, error)
	CreateAddress(ctx context.Context, req *userproto.CreateAddressRequest) (*userproto.AddressInfo, error)
	ListAddresses(ctx context.Context, req *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error)
	UpdateAddress(ctx context.Context, req *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error)
	DeleteAddress(ctx context.Context, req *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error)
	ClaimCoupon(ctx context.Context, req *userproto.ClaimCouponRequest) (*userproto.ClaimCouponResponse, error)
	ListMyCoupons(ctx context.Context, req *userproto.ListMyCouponsRequest) (*userproto.ListMyCouponsResponse, error)
	LockUserCoupon(ctx context.Context, req *userproto.LockUserCouponRequest) (*userproto.LockUserCouponResponse, error)
	ReleaseUserCoupon(ctx context.Context, req *userproto.ReleaseUserCouponRequest) (*userproto.ReleaseUserCouponResponse, error)
	GetWallet(ctx context.Context, req *userproto.GetWalletRequest) (*userproto.GetWalletResponse, error)
	RechargeWallet(ctx context.Context, req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error)
	WithdrawWallet(ctx context.Context, req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error)
}

// OrderClient 定义 passenger API 调用 ordersvc 的 RPC 契约。
type OrderClient interface {
	CreateOrder(ctx context.Context, req *orderproto.CreateOrderRequest) (*orderproto.CreateOrderResponse, error)
	CancelOrder(ctx context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error)
	GetOrder(ctx context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error)
	ListOrders(ctx context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error)
}

// PriceClient 定义 passenger API 调用 pricesvc 的价格预估契约。
type PriceClient interface {
	EstimatePrice(ctx context.Context, req *priceclient.EstimatePriceRequest) (*priceclient.EstimatePriceResponse, error)
	CalculateDiscount(ctx context.Context, req *priceclient.CalculateDiscountRequest) (*priceclient.CalculateDiscountResponse, error)
}

// PayClient 定义 passenger API 调用 paysvc 创建支付单的 RPC 契约。
type PayClient interface {
	CreatePayment(ctx context.Context, req *payproto.CreatePaymentRequest) (*payproto.CreatePaymentResponse, error)
	GetPayment(ctx context.Context, req *payproto.GetPaymentRequest) (*payproto.GetPaymentResponse, error)
}

// DispatchClient 定义 passenger API 查询 dispatchsvc 派单结果的 RPC 契约。
type DispatchClient interface {
	ListDispatchRecords(ctx context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error)
}

// LocationClient 定义 passenger API 调用位置服务（逆地理回显/POI 检索/司机位置/路径规划）的 RPC 契约。
type LocationClient interface {
	ReverseGeocode(ctx context.Context, req *locationproto.ReverseGeocodeReq, opts ...grpc.CallOption) (*locationproto.ReverseGeocodeResp, error)
	POISearch(ctx context.Context, req *locationproto.POISearchReq, opts ...grpc.CallOption) (*locationproto.POISearchResp, error)
	GetDriverLocation(ctx context.Context, req *locationproto.GetDriverLocationReq, opts ...grpc.CallOption) (*locationproto.GetDriverLocationResp, error)
	RoutePlan(ctx context.Context, req *locationproto.RoutePlanReq, opts ...grpc.CallOption) (*locationproto.RoutePlanResp, error)
}

// Option 用于在本地联调和测试时按需注入下游客户端与配置。
type Option func(*ServiceContext)

// ServiceContext 保存 passenger API 运行时依赖。
type ServiceContext struct {
	UserClient      UserClient
	OrderClient     OrderClient
	PriceClient     PriceClient
	PayClient       PayClient
	DispatchClient  DispatchClient
	LocationClient  LocationClient
	Reviews         ReviewRepository
	TokenSigningKey string
	PriceCityCode   string
	Qiniu           *qiniuutil.Client
	grpcConns       []*grpc.ClientConn
}

// NewServiceContext 创建 passenger API 运行时依赖集合。
func NewServiceContext(userClient UserClient, opts ...Option) *ServiceContext {
	ctx := &ServiceContext{
		UserClient:    userClient,
		PriceCityCode: defaultPriceCityCode,
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// LoadRuntimeConfig 从 YAML 配置文件加载 passenger API 配置，再叠加环境变量覆盖。
// 配置文件不存在时返回明确错误，避免手动启动时悄悄漏掉 JWT 密钥或 RPC 地址。
func LoadRuntimeConfig(configFile string) (RuntimeConfig, error) {
	// 本地开发时自动加载 YAML 同目录下的 .env，密钥无需手动注入 shell。
	if err := loadDotEnvFile(filepath.Join(filepath.Dir(configFile), ".env")); err != nil {
		return RuntimeConfig{}, err
	}
	var cfg RuntimeConfig
	if err := conf.LoadConfig(configFile, &cfg); err != nil {
		return RuntimeConfig{}, err
	}
	return applyRuntimeDefaults(applyRuntimeEnvOverrides(cfg)), nil
}

// LoadRuntimeConfigFromEnv 从环境变量加载 passenger API 配置，主要供单元测试或临时脚本复用。
// PASSENGER_* 变量缺省时使用本地默认 gRPC 地址，确保启动后连接真实微服务。
func LoadRuntimeConfigFromEnv() RuntimeConfig {
	return applyRuntimeDefaults(applyRuntimeEnvOverrides(RuntimeConfig{}))
}

// applyRuntimeEnvOverrides 使用环境变量覆盖 YAML 配置，便于本地临时切换端口或下游 RPC 地址。
func applyRuntimeEnvOverrides(cfg RuntimeConfig) RuntimeConfig {
	cfg.HTTPAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_HTTP_ADDR"), cfg.HTTPAddr)
	cfg.TokenSigningKey = firstNonEmptyRuntime(os.Getenv("PASSENGER_TOKEN_SIGNING_KEY"), cfg.TokenSigningKey)
	cfg.UserRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_USERSVC_ADDR"), cfg.UserRPCAddr)
	cfg.OrderRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_ORDERSVC_ADDR"), cfg.OrderRPCAddr)
	cfg.PriceRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_PRICESVC_ADDR"), cfg.PriceRPCAddr)
	cfg.PayRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_PAYSVC_ADDR"), cfg.PayRPCAddr)
	cfg.DispatchRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_DISPATCHSVC_ADDR"), cfg.DispatchRPCAddr)
	cfg.LocationRPCAddr = firstNonEmptyRuntime(os.Getenv("PASSENGER_LOCATIONSVC_ADDR"), cfg.LocationRPCAddr)
	cfg.ClientMode = firstNonEmptyRuntime(os.Getenv("PASSENGER_CLIENT_MODE"), cfg.ClientMode)
	cfg.PriceCityCode = firstNonEmptyRuntime(os.Getenv("PASSENGER_PRICE_CITY_CODE"), cfg.PriceCityCode)
	cfg.MysqlDSN = firstNonEmptyRuntime(os.Getenv("PASSENGER_MYSQL_DSN"), cfg.MysqlDSN)
	cfg.QiniuAccessKey = firstNonEmptyRuntime(os.Getenv("PASSENGER_QINIU_ACCESS_KEY"), cfg.QiniuAccessKey)
	cfg.QiniuSecretKey = firstNonEmptyRuntime(os.Getenv("PASSENGER_QINIU_SECRET_KEY"), cfg.QiniuSecretKey)
	cfg.QiniuBucket = firstNonEmptyRuntime(os.Getenv("PASSENGER_QINIU_BUCKET"), cfg.QiniuBucket)
	cfg.QiniuDomain = firstNonEmptyRuntime(os.Getenv("PASSENGER_QINIU_DOMAIN"), cfg.QiniuDomain)
	cfg.QiniuUploadURL = firstNonEmptyRuntime(os.Getenv("PASSENGER_QINIU_UPLOAD_URL"), cfg.QiniuUploadURL)
	return cfg
}

// firstNonEmptyRuntime 返回第一个非空配置值，并统一清理首尾空白。
func firstNonEmptyRuntime(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// NewServiceContextFromConfig 按配置创建 ServiceContext。
// 未提供 RPC 地址时使用默认 gRPC 地址，禁止静默回退到本地内存客户端。
func NewServiceContextFromConfig(cfg RuntimeConfig) (*ServiceContext, error) {
	cfg = applyRuntimeDefaults(cfg)
	if strings.TrimSpace(cfg.TokenSigningKey) == "" {
		return nil, fmt.Errorf("passenger token signing key is required")
	}

	if cfg.ClientMode == clientModeLocal {
		return newLocalServiceContext(cfg), nil
	}
	if cfg.ClientMode != clientModeGRPC {
		return nil, fmt.Errorf("unsupported passenger client mode: %s", cfg.ClientMode)
	}

	userClient, userConn, err := buildUserClient(cfg.UserRPCAddr)
	if err != nil {
		return nil, err
	}
	orderClient, orderConn, err := buildOrderClient(cfg.OrderRPCAddr)
	if err != nil {
		closeGRPCConns(userConn)
		return nil, err
	}
	priceClient, priceConn, err := buildPriceClient(cfg.PriceRPCAddr, cfg.PriceCityCode)
	if err != nil {
		closeGRPCConns(userConn, orderConn)
		return nil, err
	}
	payClient, payConn, err := buildPayClient(cfg.PayRPCAddr)
	if err != nil {
		closeGRPCConns(userConn, orderConn, priceConn)
		return nil, err
	}
	dispatchClient, dispatchConn, err := buildDispatchClient(cfg.DispatchRPCAddr)
	if err != nil {
		closeGRPCConns(userConn, orderConn, priceConn, payConn)
		return nil, err
	}
	locationClient, locationConn, err := buildLocationClient(cfg.LocationRPCAddr)
	if err != nil {
		closeGRPCConns(userConn, orderConn, priceConn, payConn, dispatchConn)
		return nil, err
	}

	ctx := NewServiceContext(
		userClient,
		WithOrderClient(orderClient),
		WithPriceClient(priceClient),
		WithPayClient(payClient),
		WithDispatchClient(dispatchClient),
		WithLocationClient(locationClient),
		WithTokenSigningKey(cfg.TokenSigningKey),
		WithPriceCityCode(cfg.PriceCityCode),
	)
	if cfg.QiniuAccessKey != "" || cfg.QiniuSecretKey != "" || cfg.QiniuBucket != "" || cfg.QiniuDomain != "" {
		qiniuClient, err := qiniuutil.NewClient(qiniuutil.Config{AccessKey: cfg.QiniuAccessKey, SecretKey: cfg.QiniuSecretKey, Bucket: cfg.QiniuBucket, Domain: cfg.QiniuDomain, UploadURL: cfg.QiniuUploadURL})
		if err != nil {
			return nil, err
		}
		ctx.Qiniu = qiniuClient
	}
	if cfg.MysqlDSN != "" {
		db, err := datasource.NewMysqlClient(commonconfig.MysqlConf{
			Dsn:         cfg.MysqlDSN,
			MaxOpenConn: 10,
			MaxIdleConn: 5,
			MaxLifeTime: int((30 * time.Minute).Nanoseconds()),
		})
		if err != nil {
			closeGRPCConns(userConn, orderConn, priceConn, payConn, dispatchConn)
			return nil, err
		}
		ctx.Reviews = NewGormReviewRepository(db)
	}
	ctx.grpcConns = compactGRPCConns(userConn, orderConn, priceConn, payConn, dispatchConn, locationConn)
	return ctx, nil
}

// newLocalServiceContext 创建显式 local 模式下的本地客户端集合，仅用于测试和无下游依赖的本地演示。
func newLocalServiceContext(cfg RuntimeConfig) *ServiceContext {
	return NewServiceContext(
		userlocal.NewLocalClient(cfg.TokenSigningKey, func(phone, code string) {
			logx.Infof("[LOCAL SMS] 手机号=%s 验证码=%s（local 模式不会真实发送短信）", phone, code)
		}),
		WithOrderClient(orderlocal.NewLocalClient()),
		WithPriceClient(priceclient.NewLocalClient()),
		WithPayClient(newLocalPayClient(paylocal.NewLocalClient())),
		WithDispatchClient(newMemoryDispatchClient()),
		WithReviewRepository(NewMemoryReviewRepository()),
		WithTokenSigningKey(cfg.TokenSigningKey),
		WithPriceCityCode(cfg.PriceCityCode),
	)
}

// applyRuntimeDefaults 补齐乘客端网关默认运行参数，默认下游均为真实 gRPC 服务地址。
func applyRuntimeDefaults(cfg RuntimeConfig) RuntimeConfig {
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}
	if cfg.TokenSigningKey == "" {
		cfg.TokenSigningKey = strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY"))
	}
	// 本地开发模式使用固定开发密钥，避免每次启动都因未配置环境变量退出。
	if strings.TrimSpace(cfg.TokenSigningKey) == "" && cfg.ClientMode == clientModeLocal {
		cfg.TokenSigningKey = localDevSigningKey
	}
	if cfg.UserRPCAddr == "" {
		cfg.UserRPCAddr = defaultUserRPCAddr
	}
	if cfg.OrderRPCAddr == "" {
		cfg.OrderRPCAddr = defaultOrderRPCAddr
	}
	if cfg.PriceRPCAddr == "" {
		cfg.PriceRPCAddr = defaultPriceRPCAddr
	}
	if cfg.PayRPCAddr == "" {
		cfg.PayRPCAddr = defaultPayRPCAddr
	}
	if cfg.DispatchRPCAddr == "" {
		cfg.DispatchRPCAddr = defaultDispatchRPCAddr
	}
	if cfg.LocationRPCAddr == "" {
		cfg.LocationRPCAddr = defaultLocationRPCAddr
	}
	if cfg.ClientMode == "" {
		// 乘客端 API 默认必须调用真实 usersvc gRPC，避免验证码接口只写入本地内存而没有真正发送短信。
		cfg.ClientMode = clientModeGRPC
	}
	if cfg.PriceCityCode == "" {
		cfg.PriceCityCode = defaultPriceCityCode
	}
	return cfg
}

// Close 关闭真实 gRPC 连接。
func (ctx *ServiceContext) Close() {
	if ctx == nil {
		return
	}
	closeGRPCConns(ctx.grpcConns...)
	ctx.grpcConns = nil
}

// WithOrderClient 注入订单服务客户端。
func WithOrderClient(client OrderClient) Option {
	return func(ctx *ServiceContext) {
		ctx.OrderClient = client
	}
}

// WithPriceClient 注入价格服务客户端。
func WithPriceClient(client PriceClient) Option {
	return func(ctx *ServiceContext) {
		ctx.PriceClient = client
	}
}

// WithPriceCityCode 设置乘客端默认城市编码，供下单请求未传城市时使用。
func WithPriceCityCode(cityCode string) Option {
	return func(ctx *ServiceContext) {
		cityCode = strings.TrimSpace(cityCode)
		if cityCode != "" {
			ctx.PriceCityCode = cityCode
		}
	}
}

// WithPayClient 注入支付服务客户端。
func WithPayClient(client PayClient) Option {
	return func(ctx *ServiceContext) {
		ctx.PayClient = client
	}
}

// WithDispatchClient 注入派单服务客户端。
func WithDispatchClient(client DispatchClient) Option {
	return func(ctx *ServiceContext) {
		ctx.DispatchClient = client
	}
}

// WithLocationClient 注入位置服务客户端。
func WithLocationClient(client LocationClient) Option {
	return func(ctx *ServiceContext) {
		ctx.LocationClient = client
	}
}

// WithReviewRepository 注入订单评价仓储，测试或本地模式可替换为内存实现。
func WithReviewRepository(repo ReviewRepository) Option {
	return func(ctx *ServiceContext) {
		ctx.Reviews = repo
	}
}

// WithTokenSigningKey 设置 JWT 解析签名密钥。
func WithTokenSigningKey(signingKey string) Option {
	return func(ctx *ServiceContext) {
		if signingKey != "" {
			ctx.TokenSigningKey = signingKey
		}
	}
}

// buildUserClient 根据 usersvc 地址创建真实 gRPC 客户端。
func buildUserClient(addr string) (UserClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("usersvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCUserClient(userproto.NewUserClient(conn)), conn, nil
}

// buildOrderClient 根据 ordersvc 地址创建真实 gRPC 客户端。
func buildOrderClient(addr string) (OrderClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("ordersvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCOrderClient(orderproto.NewOrderClient(conn)), conn, nil
}

// buildPriceClient 根据 pricesvc 地址创建真实 gRPC 客户端。
func buildPriceClient(addr, cityCode string) (PriceClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("pricesvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCPriceClient(conn, cityCode), conn, nil
}

// buildPayClient 根据 paysvc 地址创建真实 gRPC 客户端。
func buildPayClient(addr string) (PayClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("paysvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCPayClient(payproto.NewPayClient(conn)), conn, nil
}

// buildDispatchClient 根据 dispatchsvc 地址创建真实 gRPC 客户端。
func buildDispatchClient(addr string) (DispatchClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("dispatchsvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCDispatchClient(dispatchproto.NewDispatchClient(conn)), conn, nil
}

// buildLocationClient 根据 locationsvc 地址创建真实 gRPC 客户端。
func buildLocationClient(addr string) (LocationClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, nil, fmt.Errorf("locationsvc grpc addr is required")
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return locationproto.NewLocationServiceClient(conn), conn, nil
}

// closeGRPCConns 关闭非空 gRPC 连接，忽略关闭阶段错误。
func closeGRPCConns(conns ...*grpc.ClientConn) {
	for _, conn := range conns {
		if conn != nil {
			_ = conn.Close()
		}
	}
}

// compactGRPCConns 过滤空连接，便于 ServiceContext 统一释放资源。
func compactGRPCConns(conns ...*grpc.ClientConn) []*grpc.ClientConn {
	out := make([]*grpc.ClientConn, 0, len(conns))
	for _, conn := range conns {
		if conn != nil {
			out = append(out, conn)
		}
	}
	return out
}

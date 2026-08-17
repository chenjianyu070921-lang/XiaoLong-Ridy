package svc

import (
	"context"
	"os"
	"strings"

	orderlocal "XiaoLong-Ridy/rpc/ordersvc/client"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	userlocal "XiaoLong-Ridy/rpc/usersvc/client"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc"
)

const (
	defaultHTTPAddr        = ":8091"
	defaultTokenSigningKey = "local-development-signing-key"
	defaultPriceCityCode   = "110000"
)

// RuntimeConfig 保存 passenger API 启动时需要的运行参数。
// RPC 地址为空时会使用本地内存客户端，便于没有中间件的本地演示。
type RuntimeConfig struct {
	HTTPAddr        string
	TokenSigningKey string
	UserRPCAddr     string
	OrderRPCAddr    string
	PriceRPCAddr    string
	PriceCityCode   string
}

// UserClient 定义 passenger API 调用 usersvc 的完整 RPC 契约。
type UserClient interface {
	SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error)
	LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error)
	RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error)
	Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error)
	GetProfile(ctx context.Context, req *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error)
	SubmitRealName(ctx context.Context, req *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error)
	CreateAddress(ctx context.Context, req *userproto.CreateAddressRequest) (*userproto.AddressInfo, error)
	ListAddresses(ctx context.Context, req *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error)
	UpdateAddress(ctx context.Context, req *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error)
	DeleteAddress(ctx context.Context, req *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error)
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

// Option 用于在本地联调和测试时按需注入下游客户端与配置。
type Option func(*ServiceContext)

// ServiceContext 保存 passenger API 运行时依赖。
type ServiceContext struct {
	UserClient      UserClient
	OrderClient     OrderClient
	PriceClient     PriceClient
	TokenSigningKey string
	grpcConns       []*grpc.ClientConn
}

// NewServiceContext 创建 passenger API 运行时依赖集合。
func NewServiceContext(userClient UserClient, opts ...Option) *ServiceContext {
	ctx := &ServiceContext{
		UserClient:      userClient,
		TokenSigningKey: "local-development-signing-key",
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// LoadRuntimeConfigFromEnv 从环境变量加载 passenger API 配置。
// PASSENGER_* 变量缺省时使用本地演示配置，不影响现有启动方式。
func LoadRuntimeConfigFromEnv() RuntimeConfig {
	cfg := RuntimeConfig{
		HTTPAddr:        strings.TrimSpace(os.Getenv("PASSENGER_HTTP_ADDR")),
		TokenSigningKey: strings.TrimSpace(os.Getenv("PASSENGER_TOKEN_SIGNING_KEY")),
		UserRPCAddr:     strings.TrimSpace(os.Getenv("PASSENGER_USERSVC_ADDR")),
		OrderRPCAddr:    strings.TrimSpace(os.Getenv("PASSENGER_ORDERSVC_ADDR")),
		PriceRPCAddr:    strings.TrimSpace(os.Getenv("PASSENGER_PRICESVC_ADDR")),
		PriceCityCode:   strings.TrimSpace(os.Getenv("PASSENGER_PRICE_CITY_CODE")),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = defaultHTTPAddr
	}
	if cfg.TokenSigningKey == "" {
		cfg.TokenSigningKey = defaultTokenSigningKey
	}
	if cfg.PriceCityCode == "" {
		cfg.PriceCityCode = defaultPriceCityCode
	}
	return cfg
}

// NewServiceContextFromConfig 按配置创建 ServiceContext。
// 未提供 RPC 地址时保留 LocalClient 回退，提供地址时注入真实 gRPC 客户端。
func NewServiceContextFromConfig(cfg RuntimeConfig) (*ServiceContext, error) {
	return NewServiceContextFromConfigWithSMSLogger(cfg, nil)
}

// NewServiceContextFromConfigWithSMSLogger 创建 ServiceContext，并允许 main 注入本地验证码日志回调。
func NewServiceContextFromConfigWithSMSLogger(cfg RuntimeConfig, onSMSCode func(phone, code string)) (*ServiceContext, error) {
	if cfg.TokenSigningKey == "" {
		cfg.TokenSigningKey = defaultTokenSigningKey
	}
	if cfg.PriceCityCode == "" {
		cfg.PriceCityCode = defaultPriceCityCode
	}

	userClient, userConn, err := buildUserClient(cfg.UserRPCAddr, cfg.TokenSigningKey, onSMSCode)
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

	ctx := NewServiceContext(
		userClient,
		WithOrderClient(orderClient),
		WithPriceClient(priceClient),
		WithTokenSigningKey(cfg.TokenSigningKey),
	)
	ctx.grpcConns = compactGRPCConns(userConn, orderConn, priceConn)
	return ctx, nil
}

// Close 关闭真实 gRPC 连接；LocalClient 回退场景没有额外资源。
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

// WithTokenSigningKey 设置 JWT 解析签名密钥。
func WithTokenSigningKey(signingKey string) Option {
	return func(ctx *ServiceContext) {
		if signingKey != "" {
			ctx.TokenSigningKey = signingKey
		}
	}
}

// buildUserClient 根据地址决定使用真实 usersvc gRPC 客户端或本地内存客户端。
func buildUserClient(addr, signingKey string, onSMSCode func(phone, code string)) (UserClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return userlocal.NewLocalClient(signingKey, onSMSCode), nil, nil
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCUserClient(userproto.NewUserClient(conn)), conn, nil
}

// buildOrderClient 根据地址决定使用真实 ordersvc gRPC 客户端或本地内存客户端。
func buildOrderClient(addr string) (OrderClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return orderlocal.NewLocalClient(), nil, nil
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCOrderClient(orderproto.NewOrderClient(conn)), conn, nil
}

// buildPriceClient 根据地址决定使用真实 pricesvc gRPC 客户端或本地计价客户端。
func buildPriceClient(addr, cityCode string) (PriceClient, *grpc.ClientConn, error) {
	if strings.TrimSpace(addr) == "" {
		return priceclient.NewLocalClient(), nil, nil
	}
	conn, err := newInsecureGRPCConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return newGRPCPriceClient(conn, cityCode), conn, nil
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

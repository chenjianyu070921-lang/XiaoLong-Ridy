package svc

import (
	"context"

	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// UserClient 定义 passenger API 调用 usersvc 的完整 RPC 契约。
type UserClient interface {
	SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error)
	LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error)
	RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error)
	Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error)
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
}

// Option 用于在本地联调和测试时按需注入下游客户端与配置。
type Option func(*ServiceContext)

// ServiceContext 保存 passenger API 运行时依赖。
type ServiceContext struct {
	UserClient      UserClient
	OrderClient     OrderClient
	PriceClient     PriceClient
	TokenSigningKey string
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

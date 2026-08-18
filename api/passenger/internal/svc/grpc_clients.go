package svc

import (
	"context"
	"math"
	"strings"
	"time"

	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	priceproto "XiaoLong-Ridy/rpc/pricesvc/proto"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcUserClient 将 usersvc 生成的 gRPC 客户端适配为 passenger 的 UserClient 接口。
type grpcUserClient struct {
	cli userproto.UserClient
}

// newGRPCUserClient 创建 usersvc gRPC adapter。
func newGRPCUserClient(cli userproto.UserClient) *grpcUserClient {
	return &grpcUserClient{cli: cli}
}

func (c *grpcUserClient) SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	return c.cli.SendSMSCode(ctx, req)
}

func (c *grpcUserClient) LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	return c.cli.LoginBySMS(ctx, req)
}

func (c *grpcUserClient) RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	return c.cli.RefreshToken(ctx, req)
}

func (c *grpcUserClient) Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	return c.cli.Logout(ctx, req)
}

func (c *grpcUserClient) GetProfile(ctx context.Context, req *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error) {
	return c.cli.GetProfile(ctx, req)
}

func (c *grpcUserClient) SubmitRealName(ctx context.Context, req *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error) {
	return c.cli.SubmitRealName(ctx, req)
}

func (c *grpcUserClient) CreateAddress(ctx context.Context, req *userproto.CreateAddressRequest) (*userproto.AddressInfo, error) {
	return c.cli.CreateAddress(ctx, req)
}

func (c *grpcUserClient) ListAddresses(ctx context.Context, req *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error) {
	return c.cli.ListAddresses(ctx, req)
}

func (c *grpcUserClient) UpdateAddress(ctx context.Context, req *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error) {
	return c.cli.UpdateAddress(ctx, req)
}

func (c *grpcUserClient) DeleteAddress(ctx context.Context, req *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error) {
	return c.cli.DeleteAddress(ctx, req)
}

func (c *grpcUserClient) ClaimCoupon(ctx context.Context, req *userproto.ClaimCouponRequest) (*userproto.ClaimCouponResponse, error) {
	return c.cli.ClaimCoupon(ctx, req)
}

func (c *grpcUserClient) ListMyCoupons(ctx context.Context, req *userproto.ListMyCouponsRequest) (*userproto.ListMyCouponsResponse, error) {
	return c.cli.ListMyCoupons(ctx, req)
}

// LockUserCoupon 调用 usersvc 锁定用户券，供下单前状态校验和防重复使用。
func (c *grpcUserClient) LockUserCoupon(ctx context.Context, req *userproto.LockUserCouponRequest) (*userproto.LockUserCouponResponse, error) {
	return c.cli.LockUserCoupon(ctx, req)
}

// ReleaseUserCoupon 调用 usersvc 释放下单失败时已锁定的用户券。
func (c *grpcUserClient) ReleaseUserCoupon(ctx context.Context, req *userproto.ReleaseUserCouponRequest) (*userproto.ReleaseUserCouponResponse, error) {
	return c.cli.ReleaseUserCoupon(ctx, req)
}

// grpcOrderClient 将 ordersvc 生成的 gRPC 客户端适配为 passenger 的 OrderClient 接口。
type grpcOrderClient struct {
	cli orderproto.OrderClient
}

// newGRPCOrderClient 创建 ordersvc gRPC adapter。
func newGRPCOrderClient(cli orderproto.OrderClient) *grpcOrderClient {
	return &grpcOrderClient{cli: cli}
}

func (c *grpcOrderClient) CreateOrder(ctx context.Context, req *orderproto.CreateOrderRequest) (*orderproto.CreateOrderResponse, error) {
	return c.cli.CreateOrder(ctx, req)
}

func (c *grpcOrderClient) CancelOrder(ctx context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	return c.cli.CancelOrder(ctx, req)
}

func (c *grpcOrderClient) GetOrder(ctx context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	return c.cli.GetOrder(ctx, req)
}

func (c *grpcOrderClient) ListOrders(ctx context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	return c.cli.ListOrders(ctx, req)
}

// grpcPayClient 将 paysvc 生成的 gRPC 客户端适配为 passenger 的 PayClient 接口。
type grpcPayClient struct {
	cli payproto.PayClient
}

// newGRPCPayClient 创建 paysvc gRPC adapter。
func newGRPCPayClient(cli payproto.PayClient) *grpcPayClient {
	return &grpcPayClient{cli: cli}
}

// CreatePayment 调用 paysvc 创建支付单。
func (c *grpcPayClient) CreatePayment(ctx context.Context, req *payproto.CreatePaymentRequest) (*payproto.CreatePaymentResponse, error) {
	return c.cli.CreatePayment(ctx, req)
}

// grpcPriceClient 将 pricesvc proto 客户端适配为 passenger 当前的价格预估接口。
type grpcPriceClient struct {
	cli      priceproto.PriceClient
	cityCode string
}

// newGRPCPriceClient 创建 pricesvc gRPC adapter。
func newGRPCPriceClient(conn grpc.ClientConnInterface, cityCode string) *grpcPriceClient {
	return newGRPCPriceClientFromProto(priceproto.NewPriceClient(conn), cityCode)
}

// newGRPCPriceClientFromProto 使用已生成的 pricesvc proto client 创建 adapter，便于单元测试注入假客户端。
func newGRPCPriceClientFromProto(cli priceproto.PriceClient, cityCode string) *grpcPriceClient {
	cityCode = strings.TrimSpace(cityCode)
	if cityCode == "" {
		cityCode = defaultPriceCityCode
	}
	return &grpcPriceClient{
		cli:      cli,
		cityCode: cityCode,
	}
}

// EstimatePrice 将 passenger 的坐标预估请求转换成 pricesvc 需要的里程、时长和城市编码。
func (c *grpcPriceClient) EstimatePrice(ctx context.Context, req *priceclient.EstimatePriceRequest) (*priceclient.EstimatePriceResponse, error) {
	distanceM := req.EstimatedMeters
	if distanceM <= 0 {
		distanceM = haversineMeters(req.FromLatitude, req.FromLongitude, req.ToLatitude, req.ToLongitude)
	}
	if distanceM <= 0 {
		distanceM = 1000
	}

	durationS := req.EstimatedSecond
	if durationS <= 0 {
		durationS = int64(math.Ceil(float64(distanceM) / 250.0))
	}
	if durationS <= 0 {
		durationS = 60
	}

	resp, err := c.cli.EstimatePrice(ctx, &priceproto.EstimatePriceRequest{
		UserId:    req.UserID,
		CityCode:  c.cityCode,
		CarType:   req.CarType,
		DistanceM: distanceM,
		DurationS: durationS,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		return nil, err
	}
	return &priceclient.EstimatePriceResponse{
		EstimatedPriceCents: resp.GetTotalCents(),
		EstimatedDistanceM:  distanceM,
		EstimatedDurationS:  durationS,
	}, nil
}

// CalculateDiscount 将 passenger 的优惠券参数转发给真实 pricesvc 计算抵扣金额。
func (c *grpcPriceClient) CalculateDiscount(ctx context.Context, req *priceclient.CalculateDiscountRequest) (*priceclient.CalculateDiscountResponse, error) {
	resp, err := c.cli.CalculateDiscount(ctx, &priceproto.CalculateDiscountRequest{
		OrderId:    req.OrderID,
		TotalCents: req.TotalCents,
		Coupon: &priceproto.Coupon{
			CouponId:         req.Coupon.CouponID,
			Type:             priceproto.CouponType(req.Coupon.Type),
			FaceValueCents:   req.Coupon.FaceValueCents,
			Discount:         req.Coupon.Discount,
			ThresholdCents:   req.Coupon.ThresholdCents,
			MaxDiscountCents: req.Coupon.MaxDiscountCents,
		},
	})
	if err != nil {
		return nil, err
	}
	return &priceclient.CalculateDiscountResponse{
		DiscountAmountCents:  resp.GetDiscountAmountCents(),
		PlatformSubsidyCents: resp.GetPlatformSubsidyCents(),
		PayableAmountCents:   resp.GetPayableAmountCents(),
	}, nil
}

// newInsecureGRPCConn 创建本地/内网联调用的明文 gRPC 连接。
func newInsecureGRPCConn(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(strings.TrimSpace(addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// haversineMeters 使用球面距离公式估算两组经纬度之间的直线距离。
func haversineMeters(lat1, lon1, lat2, lon2 float64) int64 {
	const earthRadius = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return int64(math.Round(earthRadius * c))
}

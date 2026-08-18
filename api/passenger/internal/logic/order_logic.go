package logic

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// OrderLogic 封装乘客端订单相关业务流程。
type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewOrderLogic 创建订单业务逻辑实例。
func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *OrderLogic {
	return &OrderLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// CreateOrder 完成“预估价格 -> 创建订单”的乘客下单流程。
func (l *OrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if err := validateCreateOrder(req); err != nil {
		return nil, err
	}
	priceClient, err := l.priceClient()
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}

	price, err := priceClient.EstimatePrice(l.ctx, &priceclient.EstimatePriceRequest{
		UserID:        int64(userID),
		CarType:       req.CarType,
		FromLongitude: req.FromLongitude,
		FromLatitude:  req.FromLatitude,
		ToLongitude:   req.ToLongitude,
		ToLatitude:    req.ToLatitude,
	})
	if err != nil {
		return nil, err
	}
	originalPriceCents := price.EstimatedPriceCents
	discountAmountCents := int64(0)
	payableAmountCents := originalPriceCents
	lockOrderID := uint64(0)
	if hasUserCoupon(req) {
		userClient, err := l.userClient()
		if err != nil {
			return nil, err
		}
		lockOrderID = couponLockOrderID()
		lockedCoupon, err := userClient.LockUserCoupon(l.ctx, &userproto.LockUserCouponRequest{
			UserId:       userID,
			UserCouponId: req.UserCouponID,
			OrderId:      lockOrderID,
			CarType:      req.CarType,
		})
		if err != nil {
			return nil, err
		}
		if lockedCoupon.GetCoupon() == nil {
			return nil, ErrInvalidRequest
		}
		discount, err := priceClient.CalculateDiscount(l.ctx, &priceclient.CalculateDiscountRequest{
			TotalCents: originalPriceCents,
			OrderID:    int64(lockOrderID),
			Coupon:     toPriceCoupon(lockedCoupon.GetCoupon()),
		})
		if err != nil {
			releaseLockedCoupon(l.ctx, userClient, userID, req.UserCouponID, lockOrderID)
			return nil, err
		}
		discountAmountCents = discount.DiscountAmountCents
		payableAmountCents = discount.PayableAmountCents
	}

	order, err := orderClient.CreateOrder(l.ctx, &orderproto.CreateOrderRequest{
		UserId:              int64(userID),
		CarType:             req.CarType,
		FromAddress:         strings.TrimSpace(req.FromAddress),
		FromLongitude:       req.FromLongitude,
		FromLatitude:        req.FromLatitude,
		ToAddress:           strings.TrimSpace(req.ToAddress),
		ToLongitude:         req.ToLongitude,
		ToLatitude:          req.ToLatitude,
		EstimatedDistanceM:  price.EstimatedDistanceM,
		EstimatedDurationS:  price.EstimatedDurationS,
		EstimatedPriceCents: payableAmountCents,
	})
	if err != nil {
		if hasUserCoupon(req) {
			userClient, clientErr := l.userClient()
			if clientErr == nil {
				releaseLockedCoupon(l.ctx, userClient, userID, req.UserCouponID, lockOrderID)
			}
		}
		return nil, err
	}
	return &types.CreateOrderResponse{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		EstimatedPriceCents: order.GetEstimatedPriceCents(),
		OriginalPriceCents:  originalPriceCents,
		DiscountAmountCents: discountAmountCents,
		PayableAmountCents:  payableAmountCents,
		UserCouponID:        req.UserCouponID,
		Status:              int32(order.GetStatus()),
		CreatedAt:           order.GetCreatedAt(),
	}, nil
}

// ListOrders 查询当前乘客自己的订单列表。
func (l *OrderLogic) ListOrders(req *types.ListOrdersRequest) (*types.ListOrdersResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if err := validateListOrders(req); err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := orderClient.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		UserId:   int64(userID),
		Status:   orderproto.OrderStatus(req.Status),
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.OrderSummary, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, types.OrderSummary{
			OrderID:             item.GetOrderId(),
			OrderNo:             item.GetOrderNo(),
			FromAddress:         item.GetFromAddress(),
			ToAddress:           item.GetToAddress(),
			Status:              int32(item.GetStatus()),
			EstimatedPriceCents: item.GetEstimatedPriceCents(),
			CreatedAt:           item.GetCreatedAt(),
		})
	}
	return &types.ListOrdersResponse{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

// GetOrder 查询当前乘客自己的订单详情。
func (l *OrderLogic) GetOrder(req *types.GetOrderRequest) (*types.OrderDetail, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	return toOrderDetail(order), nil
}

// CancelOrder 取消当前乘客自己的订单。
func (l *OrderLogic) CancelOrder(req *types.CancelOrderRequest) (*types.CancelOrderResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	current, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if current.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	resp, err := orderClient.CancelOrder(l.ctx, &orderproto.CancelOrderRequest{
		OrderId:      req.OrderID,
		OperatorType: "user",
		OperatorId:   int64(userID),
		Reason:       strings.TrimSpace(req.Reason),
	})
	if err != nil {
		return nil, err
	}
	return &types.CancelOrderResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// PayOrder 校验当前乘客订单并调用 paysvc 创建支付单。
func (l *OrderLogic) PayOrder(req *types.PayOrderRequest) (*types.PayOrderResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	channel, err := toPayChannel(req.Channel)
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	payClient, err := l.payClient()
	if err != nil {
		return nil, err
	}

	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	if order.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY {
		return nil, ErrOrderNotPayable
	}
	if order.GetEstimatedPriceCents() <= 0 {
		return nil, ErrInvalidRequest
	}

	payment, err := payClient.CreatePayment(l.ctx, &payproto.CreatePaymentRequest{
		OrderId:     req.OrderID,
		UserId:      int64(userID),
		AmountCents: order.GetEstimatedPriceCents(),
		Channel:     channel,
	})
	if err != nil {
		return nil, err
	}
	return &types.PayOrderResponse{
		PaymentID:     payment.GetPaymentId(),
		PaymentNo:     payment.GetPaymentNo(),
		TransactionID: payment.GetTransactionId(),
		PayParams:     payment.GetPayParams(),
		Status:        payment.GetStatus(),
	}, nil
}

// validateCreateOrder 校验下单请求中的必填地址和车型参数。
func validateCreateOrder(req *types.CreateOrderRequest) error {
	if req == nil || strings.TrimSpace(req.FromAddress) == "" || strings.TrimSpace(req.ToAddress) == "" {
		return ErrInvalidRequest
	}
	if req.CarType < 1 || req.CarType > 3 {
		return ErrInvalidRequest
	}
	if !isValidLongitudeLatitude(req.FromLongitude, req.FromLatitude) ||
		!isValidLongitudeLatitude(req.ToLongitude, req.ToLatitude) {
		return ErrInvalidRequest
	}
	return nil
}

// hasUserCoupon 判断本次下单是否使用“用户已领取的券”，不信任前端传入的券模板面额。
func hasUserCoupon(req *types.CreateOrderRequest) bool {
	return req != nil && req.UserCouponID > 0
}

// couponLockOrderID 生成下单前锁券使用的临时锁标识，后续释放锁必须携带同一个值。
func couponLockOrderID() uint64 {
	return uint64(time.Now().UnixNano())
}

// toPriceCoupon 将 usersvc 校验并锁定后的券信息转换为 pricesvc 抵扣计算参数。
func toPriceCoupon(coupon *userproto.CouponInfo) priceclient.Coupon {
	if coupon == nil {
		return priceclient.Coupon{}
	}
	return priceclient.Coupon{
		CouponID:       int64(coupon.GetCouponId()),
		Type:           coupon.GetType(),
		FaceValueCents: coupon.GetFaceValueCents(),
		Discount:       coupon.GetDiscount(),
		ThresholdCents: coupon.GetThresholdCents(),
	}
}

// releaseLockedCoupon 在订单创建或抵扣计算失败时释放已锁定用户券，避免用户券长期卡在锁定状态。
func releaseLockedCoupon(ctx context.Context, userClient svc.UserClient, userID, userCouponID, lockOrderID uint64) {
	if userClient == nil || userID == 0 || userCouponID == 0 || lockOrderID == 0 {
		return
	}
	_, _ = userClient.ReleaseUserCoupon(ctx, &userproto.ReleaseUserCouponRequest{
		UserId:       userID,
		UserCouponId: userCouponID,
		OrderId:      lockOrderID,
	})
}

// validateListOrders 校验订单列表筛选和分页参数，0 页码/页大小交给下游按默认值归一化。
func validateListOrders(req *types.ListOrdersRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}
	if req.Status < 0 || req.Status > int32(orderproto.OrderStatus_ORDER_STATUS_CANCELLED) {
		return ErrInvalidRequest
	}
	if req.Page < 0 || req.PageSize < 0 || req.PageSize > 100 {
		return ErrInvalidRequest
	}
	return nil
}

// orderClient 获取订单服务客户端，避免业务方法重复判断空依赖。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}

// priceClient 获取价格服务客户端，供下单前预估价格使用。
func (l *OrderLogic) priceClient() (svc.PriceClient, error) {
	if l.svcCtx == nil || l.svcCtx.PriceClient == nil {
		return nil, ErrPriceClientNotConfigured
	}
	return l.svcCtx.PriceClient, nil
}

// userClient 获取用户服务客户端，供下单前锁定用户券使用。
func (l *OrderLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}

// payClient 获取支付服务客户端，供订单支付入口创建支付单使用。
func (l *OrderLogic) payClient() (svc.PayClient, error) {
	if l.svcCtx == nil || l.svcCtx.PayClient == nil {
		return nil, ErrPayClientNotConfigured
	}
	return l.svcCtx.PayClient, nil
}

// toPayChannel 将乘客端数字渠道转换为 paysvc proto 枚举。
func toPayChannel(channel int32) (payproto.PayChannel, error) {
	switch payproto.PayChannel(channel) {
	case payproto.PayChannel_PAY_CHANNEL_WECHAT,
		payproto.PayChannel_PAY_CHANNEL_ALIPAY,
		payproto.PayChannel_PAY_CHANNEL_BALANCE:
		return payproto.PayChannel(channel), nil
	default:
		return payproto.PayChannel_PAY_CHANNEL_UNSPECIFIED, ErrInvalidRequest
	}
}

// toOrderDetail 将 ordersvc 的订单详情响应转换为乘客端 API 响应结构。
func toOrderDetail(order *orderproto.GetOrderResponse) *types.OrderDetail {
	return &types.OrderDetail{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		UserID:              order.GetUserId(),
		DriverID:            order.GetDriverId(),
		CarType:             order.GetCarType(),
		FromAddress:         order.GetFromAddress(),
		FromLongitude:       order.GetFromLongitude(),
		FromLatitude:        order.GetFromLatitude(),
		ToAddress:           order.GetToAddress(),
		ToLongitude:         order.GetToLongitude(),
		ToLatitude:          order.GetToLatitude(),
		EstimatedDistanceM:  order.GetEstimatedDistanceM(),
		EstimatedDurationS:  order.GetEstimatedDurationS(),
		EstimatedPriceCents: order.GetEstimatedPriceCents(),
		Status:              int32(order.GetStatus()),
		CancelReason:        order.GetCancelReason(),
		CancelBy:            order.GetCancelBy(),
		CreatedAt:           order.GetCreatedAt(),
		UpdatedAt:           order.GetUpdatedAt(),
	}
}

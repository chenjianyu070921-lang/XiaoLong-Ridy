package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
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
		UserID:          int64(userID),
		CityCode:        l.orderCityCode(req.CityCode),
		CarType:         req.CarType,
		FromLongitude:   req.FromLongitude,
		FromLatitude:    req.FromLatitude,
		ToLongitude:     req.ToLongitude,
		ToLatitude:      req.ToLatitude,
		EstimatedMeters: req.EstimatedDistanceM,
		EstimatedSecond: req.EstimatedDurationS,
	})
	if err != nil {
		return nil, err
	}
	originalPriceCents := price.EstimatedPriceCents
	discountAmountCents := int64(0)
	payableAmountCents := originalPriceCents
	var selectedCoupon *userproto.CouponInfo
	if hasUserCoupon(req) {
		userClient, err := l.userClient()
		if err != nil {
			return nil, err
		}
		selectedCoupon, err = l.findUserCoupon(userClient, userID, req.UserCouponID)
		if err != nil {
			return nil, err
		}
		discount, err := priceClient.CalculateDiscount(l.ctx, &priceclient.CalculateDiscountRequest{
			TotalCents: originalPriceCents,
			Coupon:     toPriceCoupon(selectedCoupon),
		})
		if err != nil {
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
		EstimatedPriceCents: originalPriceCents,
		CityCode:            l.orderCityCode(req.CityCode),
		CouponId:            int64(req.UserCouponID),
		DiscountCents:       discountAmountCents,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateOrderResponse{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		EstimatedPriceCents: payableAmountCents,
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
	detail := toOrderDetail(order)
	if detail.CouponID > 0 {
		detail.CouponName = l.findCouponName(userID, uint64(detail.CouponID))
	}
	return detail, nil
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
	// 取消成功后释放该订单锁定的优惠券；没有使用优惠券时保持幂等。
	if userClient, userErr := l.userClient(); userErr == nil {
		_, _ = userClient.ReleaseUserCoupon(l.ctx, &userproto.ReleaseUserCouponRequest{UserId: userID, OrderId: uint64(req.OrderID)})
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

// GetPaymentStatus 主动查询当前乘客订单对应的支付状态。
func (l *OrderLogic) GetPaymentStatus(req *types.PaymentStatusRequest) (*types.PaymentStatusResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || (strings.TrimSpace(req.PaymentNo) == "" && req.OrderID <= 0) {
		return nil, ErrInvalidRequest
	}
	payClient, err := l.payClient()
	if err != nil {
		return nil, err
	}
	payment, err := payClient.GetPayment(l.ctx, &payproto.GetPaymentRequest{
		PaymentNo: strings.TrimSpace(req.PaymentNo),
		OrderId:   req.OrderID,
	})
	if err != nil {
		return nil, err
	}
	if payment.GetOrderId() <= 0 {
		return nil, ErrInvalidRequest
	}
	if err := l.ensureOrderOwner(payment.GetOrderId(), userID); err != nil {
		return nil, err
	}
	return toPaymentStatusResponse(payment), nil
}

// GetDispatchStatus 主动查询当前乘客订单对应的派单记录。
func (l *OrderLogic) GetDispatchStatus(req *types.DispatchStatusRequest) (*types.DispatchStatusResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	order, err := l.getOwnedOrder(req.OrderID, userID)
	if err != nil {
		return nil, err
	}
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	resp, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		OrderId:  req.OrderID,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		return nil, err
	}
	return toDispatchStatusResponse(req.OrderID, order.GetDriverId(), resp), nil
}

// validateCreateOrder 校验下单请求中的必填地址和车型参数。
func validateCreateOrder(req *types.CreateOrderRequest) error {
	if req == nil || strings.TrimSpace(req.FromAddress) == "" || strings.TrimSpace(req.ToAddress) == "" {
		return ErrInvalidRequest
	}
	if req.CouponMaxDiscountCents < 0 {
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

// findCouponName 根据订单保存的用户券实例 ID 查询优惠券名称；查询失败不影响订单详情主流程。
func (l *OrderLogic) findCouponName(userID, userCouponID uint64) string {
	userClient, err := l.userClient()
	if err != nil {
		return ""
	}
	coupons, err := userClient.ListMyCoupons(l.ctx, &userproto.ListMyCouponsRequest{
		UserId:   userID,
		Status:   0,
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		return ""
	}
	for _, coupon := range coupons.GetList() {
		if coupon.GetUserCouponId() == userCouponID {
			return strings.TrimSpace(coupon.GetName())
		}
	}
	return ""
}

// toPriceCoupon 将 usersvc 校验后的券信息转换为 pricesvc 抵扣计算参数。
// 最大抵扣额只应来自后端券模板；当前 usersvc 尚未暴露该字段，因此这里不采信前端透传值。
func toPriceCoupon(coupon *userproto.CouponInfo) priceclient.Coupon {
	if coupon == nil {
		return priceclient.Coupon{}
	}
	couponType := coupon.GetType()
	if couponType == 3 {
		// usersvc 中的 3 表示新人立减券，pricesvc 使用 1 表示固定金额立减券。
		couponType = 1
	}
	return priceclient.Coupon{
		CouponID:       int64(coupon.GetCouponId()),
		Type:           couponType,
		FaceValueCents: coupon.GetFaceValueCents(),
		Discount:       coupon.GetDiscount(),
		ThresholdCents: coupon.GetThresholdCents(),
	}
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

// dispatchClient 获取派单服务客户端，供乘客主动查询派单进展使用。
func (l *OrderLogic) dispatchClient() (svc.DispatchClient, error) {
	if l.svcCtx == nil || l.svcCtx.DispatchClient == nil {
		return nil, ErrDispatchClientNotConfigured
	}
	return l.svcCtx.DispatchClient, nil
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
		CouponID:            order.GetCouponId(),
		DiscountCents:       order.GetDiscountCents(),
		PayableCents:        order.GetPayableCents(),
		PaidCents:           order.GetPaidCents(),
		RefundCents:         order.GetRefundCents(),
		Status:              int32(order.GetStatus()),
		CancelReason:        order.GetCancelReason(),
		CancelBy:            order.GetCancelBy(),
		CreatedAt:           order.GetCreatedAt(),
		UpdatedAt:           order.GetUpdatedAt(),
	}
}

// cancelCreatedOrder cancels an order that has been created but cannot finish the coupon flow.
func cancelCreatedOrder(ctx context.Context, orderClient svc.OrderClient, orderID int64, userID uint64, reason string) {
	if orderClient == nil || orderID <= 0 || userID == 0 {
		return
	}
	_, _ = orderClient.CancelOrder(ctx, &orderproto.CancelOrderRequest{
		OrderId:      orderID,
		OperatorType: "system",
		OperatorId:   int64(userID),
		Reason:       reason,
	})
}

// findUserCoupon reads the passenger coupon before order creation so discount can be calculated without a fake order ID.
func (l *OrderLogic) findUserCoupon(userClient svc.UserClient, userID, userCouponID uint64) (*userproto.CouponInfo, error) {
	if userClient == nil || userID == 0 || userCouponID == 0 {
		return nil, ErrInvalidRequest
	}
	resp, err := userClient.ListMyCoupons(l.ctx, &userproto.ListMyCouponsRequest{
		UserId: userID,
		Status: 1,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range resp.GetList() {
		if item.GetUserCouponId() == userCouponID {
			return item, nil
		}
	}
	return nil, userproto.ErrUserCouponNotFound
}

// PollOrderStatus provides the polling fallback for passenger order status refresh.
func (l *OrderLogic) PollOrderStatus(req *types.OrderStatusPollRequest) (*types.OrderStatusPollResponse, error) {
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
	status := int32(order.GetStatus())
	resp := &types.OrderStatusPollResponse{
		OrderID:   order.GetOrderId(),
		Status:    status,
		Changed:   req.KnownStatus != status,
		UpdatedAt: order.GetUpdatedAt(),
		DriverID:  order.GetDriverId(),
	}
	if payment, err := l.paymentStatusByOrder(order.GetOrderId(), userID); err == nil {
		resp.Payment = payment
	}
	if dispatch, err := l.dispatchStatusByOrder(order, userID); err == nil {
		resp.Dispatch = dispatch
	}
	return resp, nil
}

// orderCityCode 返回请求城市编码；请求未传时使用 passenger 运行配置中的默认城市。
func (l *OrderLogic) orderCityCode(cityCode string) string {
	cityCode = strings.TrimSpace(cityCode)
	if cityCode != "" {
		return cityCode
	}
	if l != nil && l.svcCtx != nil && strings.TrimSpace(l.svcCtx.PriceCityCode) != "" {
		return strings.TrimSpace(l.svcCtx.PriceCityCode)
	}
	return "110000"
}

// ensureOrderOwner 校验订单归属，防止通过支付单号查询到其他乘客的支付信息。
func (l *OrderLogic) ensureOrderOwner(orderID int64, userID uint64) error {
	_, err := l.getOwnedOrder(orderID, userID)
	return err
}

// getOwnedOrder 查询订单并校验属于当前乘客。
func (l *OrderLogic) getOwnedOrder(orderID int64, userID uint64) (*orderproto.GetOrderResponse, error) {
	if orderID <= 0 || userID == 0 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return nil, err
	}
	if order.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	return order, nil
}

// paymentStatusByOrder 按订单 ID 查询支付状态，供订单轮询接口合并展示。
func (l *OrderLogic) paymentStatusByOrder(orderID int64, userID uint64) (*types.PaymentStatusResponse, error) {
	payClient, err := l.payClient()
	if err != nil {
		return nil, err
	}
	payment, err := payClient.GetPayment(l.ctx, &payproto.GetPaymentRequest{OrderId: orderID})
	if err != nil {
		return nil, err
	}
	if payment.GetOrderId() <= 0 {
		return nil, ErrInvalidRequest
	}
	if err := l.ensureOrderOwner(payment.GetOrderId(), userID); err != nil {
		return nil, err
	}
	return toPaymentStatusResponse(payment), nil
}

// dispatchStatusByOrder 查询派单记录，供订单轮询接口在推送断开时兜底展示候选司机。
func (l *OrderLogic) dispatchStatusByOrder(order *orderproto.GetOrderResponse, userID uint64) (*types.DispatchStatusResponse, error) {
	if order == nil || order.GetOrderId() <= 0 || order.GetUserId() != int64(userID) {
		return nil, ErrInvalidRequest
	}
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	resp, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		OrderId:  order.GetOrderId(),
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		return nil, err
	}
	return toDispatchStatusResponse(order.GetOrderId(), order.GetDriverId(), resp), nil
}

// toPaymentStatusResponse 将 paysvc 支付查询响应转换为乘客端 HTTP 响应。
func toPaymentStatusResponse(payment *payproto.GetPaymentResponse) *types.PaymentStatusResponse {
	if payment == nil {
		return nil
	}
	return &types.PaymentStatusResponse{
		PaymentID:         payment.GetPaymentId(),
		PaymentNo:         payment.GetPaymentNo(),
		OrderID:           payment.GetOrderId(),
		AmountCents:       payment.GetAmountCents(),
		Channel:           payment.GetChannel(),
		Status:            payment.GetStatus(),
		TransactionID:     payment.GetTransactionId(),
		RefundAmountCents: payment.GetRefundAmountCents(),
	}
}

// toDispatchStatusResponse 将 dispatchsvc 派单记录响应转换为乘客端 HTTP 响应。
func toDispatchStatusResponse(orderID, orderDriverID int64, resp *dispatchproto.ListDispatchRecordsResponse) *types.DispatchStatusResponse {
	result := &types.DispatchStatusResponse{
		OrderID:  orderID,
		DriverID: orderDriverID,
		Records:  make([]types.DispatchRecord, 0),
	}
	if resp == nil {
		return result
	}
	result.Total = resp.GetTotal()
	for _, item := range resp.GetList() {
		record := types.DispatchRecord{
			ID:           item.GetId(),
			OrderID:      item.GetOrderId(),
			DriverID:     item.GetDriverId(),
			DispatchType: item.GetDispatchType(),
			Status:       item.GetStatus(),
			MatchScore:   item.GetMatchScore(),
			Remark:       item.GetRemark(),
			CreatedAt:    item.GetCreatedAt(),
			UpdatedAt:    item.GetUpdatedAt(),
		}
		result.Records = append(result.Records, record)
		if shouldUseDispatchDriver(result.DriverID, record.Status, record.DriverID) {
			result.DriverID = record.DriverID
			result.DispatchStatus = record.Status
		}
	}
	return result
}

// shouldUseDispatchDriver 判断派单记录是否可作为乘客端展示的当前候选司机。
func shouldUseDispatchDriver(currentDriverID int64, status int32, driverID int64) bool {
	if driverID <= 0 {
		return false
	}
	if status == 2 {
		return true
	}
	return currentDriverID <= 0 && status == 1
}

// EstimateOrder 提供下单前实时行程费用预估，不创建订单或占用优惠券。
func (l *OrderLogic) EstimateOrder(req *types.EstimateOrderRequest) (*types.EstimateOrderResponse, error) {
	if req == nil || strings.TrimSpace(req.FromAddress) == "" || strings.TrimSpace(req.ToAddress) == "" {
		return nil, ErrInvalidRequest
	}
	if req.CarType < 1 || req.CarType > 3 {
		return nil, ErrInvalidRequest
	}
	if !isValidLongitudeLatitude(req.FromLongitude, req.FromLatitude) ||
		!isValidLongitudeLatitude(req.ToLongitude, req.ToLatitude) {
		return nil, ErrInvalidRequest
	}
	priceClient, err := l.priceClient()
	if err != nil {
		return nil, err
	}
	price, err := priceClient.EstimatePrice(l.ctx, &priceclient.EstimatePriceRequest{
		CityCode:        l.orderCityCode(req.CityCode),
		CarType:         req.CarType,
		FromLongitude:   req.FromLongitude,
		FromLatitude:    req.FromLatitude,
		ToLongitude:     req.ToLongitude,
		ToLatitude:      req.ToLatitude,
		EstimatedMeters: req.EstimatedDistanceM,
		EstimatedSecond: req.EstimatedDurationS,
	})
	if err != nil {
		return nil, err
	}

	originalPriceCents := price.EstimatedPriceCents
	discountAmountCents := int64(0)
	payableAmountCents := originalPriceCents
	if req.UserCouponID > 0 {
		userID, err := currentUserID(l.svcCtx, l.token)
		if err != nil {
			return nil, err
		}
		userClient, err := l.userClient()
		if err != nil {
			return nil, err
		}
		selectedCoupon, err := l.findUserCoupon(userClient, userID, req.UserCouponID)
		if err != nil {
			return nil, err
		}
		discount, err := priceClient.CalculateDiscount(l.ctx, &priceclient.CalculateDiscountRequest{
			TotalCents: originalPriceCents,
			Coupon:     toPriceCoupon(selectedCoupon),
		})
		if err != nil {
			return nil, err
		}
		discountAmountCents = discount.DiscountAmountCents
		payableAmountCents = discount.PayableAmountCents
	}

	return &types.EstimateOrderResponse{
		CarType:             req.CarType,
		EstimatedDistanceM:  price.EstimatedDistanceM,
		EstimatedDurationS:  price.EstimatedDurationS,
		OriginalPriceCents:  originalPriceCents,
		DiscountAmountCents: discountAmountCents,
		PayableAmountCents:  payableAmountCents,
	}, nil
}

package logic

import (
	"XiaoLong-Ridy/rpc/paysvc/pay"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/zeromicro/go-zero/core/logx"
)

// priceToleranceServer 服务端实时计价时司机上报金额的偏差容忍度（%）。
const priceToleranceServer = 10

// priceToleranceEstimate 降级为下单预估快照时司机上报金额的偏差容忍度（%）。
const priceToleranceEstimate = 50

// ErrPriceMismatch 表示司机上报实付金额与预估金额偏差过大。
var ErrPriceMismatch = errors.New("price mismatch")

type FinishTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewFinishTripLogic 创建结束行程逻辑对象。
func NewFinishTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishTripLogic {
	return &FinishTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishTrip 将行程中订单改为待支付，并对司机上报费用做金额校验。
func (l *FinishTripLogic) FinishTrip(in *proto.FinishTripRequest) (*proto.FinishTripResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 ||
		in.ActualDistanceM < 0 || in.ActualDurationS < 0 || in.ActualPriceCents < 0 {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if !CanTransit(order.Status, constants.OrderStatusWaitPay) {
		return nil, ErrOrderStatusNotAllowed
	}
	if order.DriverId != uint64(in.DriverId) {
		return nil, ErrDriverNotMatched
	}

	// 服务端计价权威：优先按实际里程/时长调 pricesvc 重算应收，失败降级为下单预估快照。
	serverAmountCents, serverPriced, settleResp := l.settleAmountCents(order, in)
	effectiveAmount := serverAmountCents
	if effectiveAmount <= 0 {
		// 无权威金额（计价服务不可用且无预估快照）时，采用司机上报值兜底。
		effectiveAmount = in.ActualPriceCents
	}
	// 司机上报金额仅做偏差校验，防止司机乱报；权威金额为服务端计价时容忍 10%，降级快照时保留 50%。
	if in.ActualPriceCents > 0 && effectiveAmount > 0 {
		tolerance := int64(priceToleranceEstimate)
		if serverPriced {
			tolerance = int64(priceToleranceServer)
		}
		diff := in.ActualPriceCents - effectiveAmount
		if diff < 0 {
			diff = -diff
		}
		if diff > effectiveAmount*tolerance/100 {
			l.Logger.Errorf("price mismatch: driver reported=%d server=%d", in.ActualPriceCents, effectiveAmount)
			return nil, ErrPriceMismatch
		}
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusWaitPay,
		OperatorType: constants.OperatorDriver,
		OperatorId:   uint64(in.DriverId),
		Remark:       fmt.Sprintf("行程结束，实际距离=%dm，实际时长=%ds，实际费用=%d分", in.ActualDistanceM, in.ActualDurationS, effectiveAmount),
	}
	ok, err := l.svcCtx.OrderRepository.FinishTrip(l.ctx, order.Id, uint64(in.DriverId), statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}

	// P1：若实际费用来自服务端实时计价，落库实际计价明细，供结算与对账使用。
	if serverPriced && l.svcCtx.PriceClient != nil && settleResp != nil {
		actual := &price.SaveActualOrderPriceRequest{
			OrderId:          int64(order.Id),
			PriceRuleId:      settleResp.PriceRuleId,
			ActualPriceCents: effectiveAmount,
			Detail:           settleResp.Detail,
		}
		if _, err := l.svcCtx.PriceClient.SaveActualOrderPrice(l.ctx, actual); err != nil {
			l.Logger.Errorf("save actual order price failed, orderId=%d: %v", order.Id, err)
		}
	}

	l.createPayment(order, effectiveAmount)

	return &proto.FinishTripResponse{
		OrderId:            in.OrderId,
		Status:             proto.OrderStatus_ORDER_STATUS_WAIT_PAY,
		PayableAmountCents: effectiveAmount,
	}, nil
}

// settleAmountCents 计算服务端权威应收金额（分）。
// 优先用实际里程/时长调 pricesvc 重算；pricesvc 未配置、调用失败或金额非法时降级为下单预估快照。
// 返回 (金额, 是否为服务端实时计价, 计价响应)；两者均无法取得时金额为 0、resp 为 nil。
func (l *FinishTripLogic) settleAmountCents(order *model.RideOrder, in *proto.FinishTripRequest) (int64, bool, *price.EstimatePriceResponse) {
	estimatedCents := int64(math.Round(order.EstimatedPrice * 100))
	if l.svcCtx.PriceClient == nil {
		return estimatedCents, false, nil
	}
	resp, err := l.svcCtx.PriceClient.EstimatePrice(l.ctx, &price.EstimatePriceRequest{
		UserId:    int64(order.UserId),
		CityCode:  orderCityCode(order.CityCode), // 优先用订单落库的真实城市，空则兜底默认城市
		CarType:   int32(order.CarType),
		DistanceM: in.ActualDistanceM,
		DurationS: in.ActualDurationS,
	})
	if err != nil || resp == nil || resp.TotalCents <= 0 {
		l.Logger.Errorf("settle price failed, fallback to estimate %d: %v", estimatedCents, err)
		return estimatedCents, false, nil
	}
	return resp.TotalCents, true, resp
}

// orderCityCode 返回订单的真实城市编码，空则兜底默认城市。
func orderCityCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return defaultCityCode
	}
	return code
}

// createPayment 调 paysvc 生成支付单，返回支付单号；失败返回空串（不阻断订单状态）。
func (l *FinishTripLogic) createPayment(order *model.RideOrder, amountCents int64) string {
	if l.svcCtx.PayClient == nil {
		l.Logger.Errorf("pay client not configured, skip create payment, orderId=%d", order.Id)
		return ""
	}
	channel := l.svcCtx.Config.PayChannel
	if channel <= 0 {
		channel = 1 // 默认微信
	}
	resp, err := l.svcCtx.PayClient.CreatePayment(l.ctx, &pay.CreatePaymentRequest{
		OrderId:     int64(order.Id),
		UserId:      int64(order.UserId),
		AmountCents: amountCents,
		Channel:     payproto.PayChannel(channel),
	})
	if err != nil || resp == nil {
		l.Logger.Errorf("create payment failed, orderId=%d: %v", order.Id, err)
		return ""
	}
	l.Logger.Infof("create payment success, orderId=%d paymentNo=%s", order.Id, resp.PaymentNo)
	return resp.PaymentNo
}

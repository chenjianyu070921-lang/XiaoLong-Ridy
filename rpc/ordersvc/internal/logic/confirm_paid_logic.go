package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"

	"github.com/zeromicro/go-zero/core/logx"
)

// paymentStatusPaid 支付成功状态值，与 paysvc 的 PaymentStatusPaid 保持一致。
const paymentStatusPaid = 2

var (
	// ErrPaymentNotConfigured 支付客户端未配置，拒绝直接完成订单。
	ErrPaymentNotConfigured = errors.New("pay client not configured")
	// ErrPaymentVerificationFailed 支付单信息与订单不一致或未支付成功。
	ErrPaymentVerificationFailed = errors.New("payment verification failed")
)

type ConfirmPaidLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewConfirmPaidLogic 创建支付成功确认逻辑。
func NewConfirmPaidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPaidLogic {
	return &ConfirmPaidLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ConfirmPaid 支付成功回调：校验支付单真实状态后，待支付 -> 已完成。
func (l *ConfirmPaidLogic) ConfirmPaid(in *proto.ConfirmPaidRequest) (*proto.ConfirmPaidResponse, error) {
	if in.OrderId <= 0 || strings.TrimSpace(in.PaymentNo) == "" || in.AmountCents <= 0 || in.PaidAt <= 0 {
		return nil, ErrInvalidOrderParams
	}
	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order.Status != constants.OrderStatusWaitPay {
		return nil, ErrOrderStatusNotAllowed
	}

	// 未配置支付客户端时拒绝直接完成订单，防止支付绕过。
	if l.svcCtx.PayClient == nil {
		return nil, ErrPaymentNotConfigured
	}
	// 向支付服务核验支付单：单号、订单、金额、状态必须完全一致。
	payment, err := l.svcCtx.PayClient.GetPayment(l.ctx, &pay.GetPaymentRequest{
		PaymentNo: strings.TrimSpace(in.PaymentNo),
		OrderId:   in.OrderId,
	})
	if err != nil {
		l.Logger.Errorf("get payment failed, orderId=%d paymentNo=%s: %v", in.OrderId, in.PaymentNo, err)
		return nil, ErrPaymentVerificationFailed
	}
	if payment == nil || payment.PaymentNo != in.PaymentNo || payment.OrderId != in.OrderId ||
		payment.AmountCents != in.AmountCents || payment.Status != paymentStatusPaid {
		l.Logger.Errorf("payment verification failed, orderId=%d paymentNo=%s, got paymentNo=%s orderId=%d amount=%d status=%d",
			in.OrderId, in.PaymentNo, payment.GetPaymentNo(), payment.GetOrderId(), payment.GetAmountCents(), payment.GetStatus())
		return nil, ErrPaymentVerificationFailed
	}

	remark := "支付成功 paymentNo=" + in.PaymentNo
	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusCompleted,
		OperatorType: constants.OperatorSystem,
		OperatorId:   0,
		Remark:       remark,
	}
	// 支付单已通过四要素核验（单号/订单/金额/状态），in.AmountCents 即乘客实付金额。
	// 必须落库 paid_cents：后续退款以 order.PaidCents 为基准，不落库会导致退款退 0 分。
	ok, err := l.svcCtx.OrderRepository.CompleteOrder(l.ctx, order.Id, statusLog, in.AmountCents)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}

	if l.svcCtx.CouponConsumer != nil {
		if err := l.svcCtx.CouponConsumer.ConsumeByOrder(l.ctx, order.UserId, order.Id); err != nil {
			return nil, err
		}
	}

	if l.svcCtx.EventBus != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"order_id":    in.OrderId,
			"from_status": constants.OrderStatusWaitPay,
			"to_status":   constants.OrderStatusCompleted,
		})
		if err := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderStatusChanged, payload); err != nil {
			l.Logger.Errorf("publish orderclient.status.changed failed: %v", err)
		}
	}

	return &proto.ConfirmPaidResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_COMPLETED,
	}, nil
}

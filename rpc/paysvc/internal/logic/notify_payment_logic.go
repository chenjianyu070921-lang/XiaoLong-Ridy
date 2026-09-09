package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/common/priceutil"
	order "XiaoLong-Ridy/rpc/ordersvc/proto"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 支付宝交易成功状态
const alipayTradeSuccess = "TRADE_SUCCESS"

// 平台抽成比例（%），默认 20%。后续可下沉到配置。
const defaultCommissionRate = 20.0

type NotifyPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotifyPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyPaymentLogic {
	return &NotifyPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 支付回调：验签 + 更新支付单 + 发事件 + 触发结算。
func (l *NotifyPaymentLogic) NotifyPayment(in *proto.NotifyPaymentRequest) (*proto.NotifyPaymentResponse, error) {
	// 1. 验签（失败直接拒绝）
	if err := l.svcCtx.Verifier.Verify(l.ctx, in.NotifyRaw); err != nil {
		return nil, fmt.Errorf("verify sign failed: %w", err)
	}

	// 2. 仅处理支付成功状态，其余状态忽略（如退款通知）
	if in.TradeStatus != alipayTradeSuccess {
		return &proto.NotifyPaymentResponse{Success: true, Message: "ignore non-success trade status"}, nil
	}

	repo := repository.NewPaymentRepo(l.svcCtx.DB)

	// 3. 查询支付单
	p, err := repo.FindByPaymentNo(l.ctx, in.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// 4. 幂等处理：已支付直接返回成功
	if p.Status == model.PaymentStatusPaid {
		return &proto.NotifyPaymentResponse{Success: true, Message: "already paid"}, nil
	}
	// 仅「待支付」状态可流转为「支付成功」
	if p.Status != model.PaymentStatusPending {
		return nil, fmt.Errorf("invalid payment status: %d", p.Status)
	}

	// 5. 更新支付单为「支付成功」
	p.Status = model.PaymentStatusPaid
	p.TransactionId = in.TransactionId
	if in.PaidAt > 0 {
		paidAt := time.Unix(in.PaidAt, 0)
		p.PaidAt = &paidAt
	}
	if err := repo.Update(l.ctx, p); err != nil {
		return nil, err
	}

	// 6. 发 Kafka「支付成功」事件（失败不阻断主流程）
	if err := l.publishPaidEvent(p, in.PaidAt); err != nil {
		l.Errorf("publish orderclient.paid event failed: %v", err)
	}

	// 6.5 通知订单服务完成订单，闭环主链路（失败不阻断回调，后续可重试）
	l.confirmOrderAfterPaid(p, in.PaidAt)

	// 7. 触发司机结算（失败不阻断主流程）
	l.settleAfterPaid(p)

	return &proto.NotifyPaymentResponse{Success: true, Message: "success"}, nil
}

// publishPaidEvent 发送支付成功事件到 Kafka。
func (l *NotifyPaymentLogic) publishPaidEvent(p *model.Payment, paidAt int64) error {
	event := &mq.OrderPaidEvent{
		OrderId:     int64(p.OrderId),
		PaymentNo:   p.PaymentNo,
		AmountCents: priceutil.YuanToCents(p.Amount),
		PaidAt:      paidAt,
	}
	data, err := mq.EncodeOrderPaidEvent(event)
	if err != nil {
		return err
	}
	return l.svcCtx.Producer.Send(constants.TopicOrderPaid, p.PaymentNo, data)
}

// confirmOrderAfterPaid 支付成功后通知订单服务确认完成订单。
func (l *NotifyPaymentLogic) confirmOrderAfterPaid(p *model.Payment, paidAt int64) {
	if paidAt <= 0 && p.PaidAt != nil {
		paidAt = p.PaidAt.Unix()
	}
	if _, err := l.svcCtx.OrderClient.ConfirmPaid(l.ctx, &order.ConfirmPaidRequest{
		OrderId:     int64(p.OrderId),
		PaymentNo:   p.PaymentNo,
		AmountCents: priceutil.YuanToCents(p.Amount),
		PaidAt:      paidAt,
	}); err != nil {
		l.Errorf("confirm order paid failed, orderId=%d paymentNo=%s: %v", p.OrderId, p.PaymentNo, err)
	}
}

// settleAfterPaid 支付成功后触发司机结算。
func (l *NotifyPaymentLogic) settleAfterPaid(p *model.Payment) {
	// 调 ordersvc 拿司机ID
	driverId, err := l.svcCtx.OrderClient.GetDriverId(l.ctx, int64(p.OrderId))
	if err != nil {
		l.Errorf("get driver_id for orderclient %d failed: %v, skip settle", p.OrderId, err)
		return
	}
	if driverId == 0 {
		l.Infof("orderclient %d has no driver, skip settle", p.OrderId)
		return
	}

	// 触发结算
	settleLogic := NewSettleOrderLogic(l.ctx, l.svcCtx)
	if _, err := settleLogic.SettleOrder(&proto.SettleOrderRequest{
		OrderId:          int64(p.OrderId),
		DriverId:         driverId,
		TotalAmountCents: priceutil.YuanToCents(p.Amount),
		CommissionRate:   defaultCommissionRate,
	}); err != nil {
		l.Errorf("settle orderclient %d failed: %v", p.OrderId, err)
	}
}

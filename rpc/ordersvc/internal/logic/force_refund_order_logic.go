package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// orderRefundedEvent 退款成功事件载荷，由 order-event-consumer 处理支付通道回款等。
type orderRefundedEvent struct {
	OrderId      int64   `json:"order_id"`
	OrderNo      string  `json:"order_no"`
	RefundNo     string  `json:"refund_no"`
	RefundCents  int64   `json:"refund_cents"`
	OperatorId   int64   `json:"operator_id"`
	OperatorType string  `json:"operator_type"`
	RefundTimeout int64  `json:"refund_timeout"`
}

type ForceRefundOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	Logger logx.Logger
}

func NewForceRefundOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForceRefundOrderLogic {
	return &ForceRefundOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ForceRefundOrder 管理员强制退款：可从更多终态（已支付/行程中/已到达等）发起，
// 状态改为已退款并累加退款金额（幂等按订单维度，不允许超已支付金额重复退款）。
func (l *ForceRefundOrderLogic) ForceRefundOrder(in *proto.ForceRefundOrderRequest) (*proto.ForceRefundOrderResponse, error) {
	if in.OrderId <= 0 || in.OperatorId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	if strings.TrimSpace(in.RefundNo) == "" {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, repository.ErrOrderNotFound
	}

	// 退款金额：未指定则按已支付金额全额退款。
	refundCents := in.RefundAmountCents
	if refundCents <= 0 {
		refundCents = order.PaidCents
	}
	if refundCents <= 0 {
		return nil, ErrInvalidOrderStatus
	}
	// 幂等保护：已退金额 + 本次退款不得超过已支付金额（防止重复/超额退款）。
	if order.RefundCents+refundCents > order.PaidCents {
		return nil, ErrRefundDuplicate
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusRefunded,
		OperatorType: constants.OperatorAdmin,
		OperatorId:   uint64(in.OperatorId),
		Remark:       "管理员强制退款：" + strings.TrimSpace(in.Reason),
		CreatedAt:    time.Now(),
	}

	updated, err := l.svcCtx.OrderRepository.ForceRefund(l.ctx, uint64(in.OrderId), refundCents, statusLog)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, ErrInvalidOrderStatus
	}

	// 释放优惠券（与正常退款一致）。
	if l.svcCtx.CouponConsumer != nil {
		if relErr := l.svcCtx.CouponConsumer.ReleaseByOrder(l.ctx, order.UserId, uint64(in.OrderId)); relErr != nil {
			l.Logger.Errorf("release coupon on force refund failed: %v", relErr)
		}
	}

	// 发布 order.refunded 事件（由 order-event-consumer 处理支付通道回款等，pay 服务未到位时仅记日志）。
	if l.svcCtx.EventBus != nil {
		payload, _ := json.Marshal(orderRefundedEvent{
			OrderId:       in.OrderId,
			OrderNo:       order.OrderNo,
			RefundNo:      in.RefundNo,
			RefundCents:   refundCents,
			OperatorId:    in.OperatorId,
			OperatorType:  constants.OperatorAdmin,
			RefundTimeout: 0,
		})
		if pubErr := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderRefunded, payload); pubErr != nil {
			l.Logger.Errorf("publish order.refunded failed: %v", pubErr)
		}
	}

	return &proto.ForceRefundOrderResponse{
		OrderId:     in.OrderId,
		Status:      constants.OrderStatusRefunded,
		RefundCents: refundCents,
	}, nil
}

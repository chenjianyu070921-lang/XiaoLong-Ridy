package logic

import (
	"context"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// refundIdemTTL 退款幂等键过期时间。
const refundIdemTTL = 24 * time.Hour

// orderRefundIdemKey 退款幂等键，防止同一 refundNo 重复退款。
func orderRefundIdemKey(refundNo string) string {
	return "order:refund:idem:" + refundNo
}

type RefundOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundOrderLogic {
	return &RefundOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefundOrderLogic) RefundOrder(in *proto.RefundOrderRequest) (*proto.RefundOrderResponse, error) {
	return l.refundOrder(in)
}

// refundOrder 已完成订单退款：状态机校验 + 幂等（refundNo） + CAS 落库 + 回滚优惠券。
func (l *RefundOrderLogic) refundOrder(in *proto.RefundOrderRequest) (*proto.RefundOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	if strings.TrimSpace(in.RefundNo) == "" {
		return nil, ErrInvalidOrderParams
	}

	// 幂等：相同 (orderId, refundNo) 不重复退款（键含订单维度，避免跨订单 refundNo 冲突）。
	idemKey := orderRefundIdemKey(in.RefundNo) + ":" + strconv.FormatInt(in.OrderId, 10)
	ok, err := l.svcCtx.Redis.SetNX(l.ctx, idemKey, 1, refundIdemTTL).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRefundDuplicate
	}

	// 读取订单，校验状态机（仅已完成 -> 已退款）。
	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if !CanTransit(order.Status, constants.OrderStatusRefunded) {
		return nil, ErrInvalidOrderStatus
	}

	// 计算退款金额：0 表示全额（以实付金额为准）。
	refundCents := in.RefundAmountCents
	if refundCents <= 0 {
		refundCents = order.PaidCents
	}
	// 幂等保护：已退金额 + 本次退款不得超过已支付金额，防止重复/超额退款。
	if order.RefundCents+refundCents > order.PaidCents {
		return nil, ErrRefundDuplicate
	}
	statusLog := &model.OrderStatusLog{
		FromStatus: order.Status,
		ToStatus:   constants.OrderStatusRefunded,
		OperatorType: constants.OperatorUser,
		OperatorId:   uint64(in.UserId),
		Remark:     in.Reason,
	}
	updated, err := l.svcCtx.OrderRepository.Refund(l.ctx, uint64(in.OrderId), refundCents, statusLog)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, ErrInvalidOrderStatus
	}

	// 回滚订单锁定的优惠券（取消/退款兜底）。
	if order.CouponId > 0 {
		if rerr := l.svcCtx.OrderRepository.ReleaseCoupon(l.ctx, uint64(in.UserId), uint64(in.OrderId)); rerr != nil {
			logx.Errorf("refund order %d release coupon %d failed: %v", in.OrderId, order.CouponId, rerr)
		}
	}

	l.Infof("order %d refunded, refundCents=%d", in.OrderId, refundCents)
	return &proto.RefundOrderResponse{
		OrderId:     in.OrderId,
		Status:      proto.OrderStatus_ORDER_STATUS_REFUNDED,
		RefundCents: refundCents,
	}, nil
}

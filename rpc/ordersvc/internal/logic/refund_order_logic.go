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

	// 幂等：相同 (orderId, refundNo) 不重复退款（键含订单维度，避免跨订单 refundNo 冲突）。
	// 放在业务校验之后：校验失败不占用键，否则同一 refundNo 在 24h TTL 内重试会被误判重复（幂等键毒化）。
	// Redis 未配置时跳过键，退化为依赖仓储层 CAS 条件更新（WHERE status=已完成）保证不重复退款。
	idemKey := orderRefundIdemKey(in.RefundNo) + ":" + strconv.FormatInt(in.OrderId, 10)
	if l.svcCtx.Redis != nil {
		ok, nxErr := l.svcCtx.Redis.SetNX(l.ctx, idemKey, 1, refundIdemTTL).Result()
		if nxErr != nil {
			return nil, nxErr
		}
		if !ok {
			return nil, ErrRefundDuplicate
		}
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
		// 落库失败释放幂等键：业务未生效不应占用键，否则重试会被误判重复。
		l.delIdemKey(idemKey)
		return nil, err
	}
	if !updated {
		l.delIdemKey(idemKey)
		return nil, ErrInvalidOrderStatus
	}

	// 回滚订单锁定的优惠券（取消/退款兜底）。
	// 用 order.UserId 而非 in.UserId：券归属订单所属用户，避免调用方传错 user 导致券释放不到、长期占用。
	if order.CouponId > 0 {
		if rerr := l.svcCtx.OrderRepository.ReleaseCoupon(l.ctx, order.UserId, uint64(in.OrderId)); rerr != nil {
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

// delIdemKey 释放退款幂等键，使业务未生效时可安全重试；失败仅记日志。
func (l *RefundOrderLogic) delIdemKey(key string) {
	if l.svcCtx == nil || l.svcCtx.Redis == nil {
		return
	}
	if err := l.svcCtx.Redis.Del(l.ctx, key).Err(); err != nil {
		l.Logger.Errorf("release refund idem key %s failed: %v", key, err)
	}
}

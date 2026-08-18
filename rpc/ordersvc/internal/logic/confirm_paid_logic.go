package logic

import (
	"context"
	"encoding/json"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
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

// ConfirmPaid 支付成功回调：待支付 -> 已完成。
func (l *ConfirmPaidLogic) ConfirmPaid(in *proto.ConfirmPaidRequest) (*proto.ConfirmPaidResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order.Status != constants.OrderStatusWaitPay {
		return nil, ErrOrderStatusNotAllowed
	}

	remark := "支付成功"
	if in.PaymentNo != "" {
		remark = "支付成功 paymentNo=" + in.PaymentNo
	}
	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusCompleted,
		OperatorType: constants.OperatorSystem,
		OperatorId:   0,
		Remark:       remark,
	}
	ok, err := l.svcCtx.OrderRepository.CompleteOrder(l.ctx, order.Id, statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
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

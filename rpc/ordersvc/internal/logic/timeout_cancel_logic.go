package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimeoutCancelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewTimeoutCancelLogic 创建超时取消逻辑对象。
func NewTimeoutCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TimeoutCancelLogic {
	return &TimeoutCancelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TimeoutCancel 由系统取消超时未接单的订单。
func (l *TimeoutCancelLogic) TimeoutCancel(in *proto.TimeoutCancelRequest) (*proto.TimeoutCancelResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "超时未接单"
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order.Status != constants.OrderStatusWaitAccept && order.Status != constants.OrderStatusAccepted {
		return nil, ErrOrderStatusNotCancelable
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusCancelled,
		OperatorType: constants.OperatorSystem,
		OperatorId:   0,
		Remark:       reason,
	}
	ok, err := l.svcCtx.OrderRepository.Cancel(l.ctx, order.Id, []int8{
		constants.OrderStatusWaitAccept,
		constants.OrderStatusAccepted,
	}, constants.OperatorSystem, reason, statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotCancelable
	}

	return &proto.TimeoutCancelResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_CANCELLED,
	}, nil
}

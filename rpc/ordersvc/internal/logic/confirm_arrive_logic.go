package logic

import (
	"context"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmArriveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewConfirmArriveLogic 创建司机到达逻辑对象。
func NewConfirmArriveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmArriveLogic {
	return &ConfirmArriveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ConfirmArrive 记录司机已到达，订单状态保持不变。
func (l *ConfirmArriveLogic) ConfirmArrive(in *proto.ConfirmArriveRequest) (*proto.ConfirmArriveResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order.Status != constants.OrderStatusAccepted {
		return nil, ErrOrderStatusNotAllowed
	}
	if order.DriverId != uint64(in.DriverId) {
		return nil, ErrDriverNotMatched
	}

	statusLog := &model.OrderStatusLog{
		OrderId:      order.Id,
		FromStatus:   order.Status,
		ToStatus:     order.Status,
		OperatorType: constants.OperatorDriver,
		OperatorId:   uint64(in.DriverId),
		Remark:       "司机已到达",
	}
	if err := l.svcCtx.OrderRepository.AppendStatusLog(l.ctx, statusLog); err != nil {
		return nil, err
	}

	return &proto.ConfirmArriveResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_ACCEPTED,
	}, nil
}

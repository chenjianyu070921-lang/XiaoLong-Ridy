package logic

import (
	"context"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type AcceptOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAcceptOrderLogic 创建接单逻辑对象。
func NewAcceptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptOrderLogic {
	return &AcceptOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AcceptOrder 将待接单订单改为已接单并绑定司机，写入接单日志。
func (l *AcceptOrderLogic) AcceptOrder(in *proto.AcceptOrderRequest) (*proto.AcceptOrderResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if !CanTransit(order.Status, constants.OrderStatusAccepted) {
		return nil, ErrOrderStatusNotAllowed
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusAccepted,
		OperatorType: constants.OperatorDriver,
		OperatorId:   uint64(in.DriverId),
		Remark:       "司机接单",
	}
	ok, err := l.svcCtx.OrderRepository.Accept(l.ctx, order.Id, uint64(in.DriverId), statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}

	return &proto.AcceptOrderResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_ACCEPTED,
	}, nil
}

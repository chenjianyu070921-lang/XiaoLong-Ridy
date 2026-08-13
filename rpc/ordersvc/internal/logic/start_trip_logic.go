package logic

import (
	"context"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewStartTripLogic 创建开始行程逻辑对象。
func NewStartTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartTripLogic {
	return &StartTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StartTrip 将已接单订单改为行程中，写入开始行程日志。
func (l *StartTripLogic) StartTrip(in *proto.StartTripRequest) (*proto.StartTripResponse, error) {
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
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusOnTrip,
		OperatorType: constants.OperatorDriver,
		OperatorId:   uint64(in.DriverId),
		Remark:       "开始行程",
	}
	ok, err := l.svcCtx.OrderRepository.StartTrip(l.ctx, order.Id, uint64(in.DriverId), statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}

	return &proto.StartTripResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_ON_TRIP,
	}, nil
}

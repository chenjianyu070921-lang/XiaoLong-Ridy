package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewFinishTripLogic 创建结束行程逻辑对象。
func NewFinishTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishTripLogic {
	return &FinishTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishTrip 将行程中订单改为待支付，并把实际费用快照写入状态日志。
func (l *FinishTripLogic) FinishTrip(in *proto.FinishTripRequest) (*proto.FinishTripResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 ||
		in.ActualDistanceM < 0 || in.ActualDurationS < 0 || in.ActualPriceCents < 0 {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if !CanTransit(order.Status, constants.OrderStatusWaitPay) {
		return nil, ErrOrderStatusNotAllowed
	}
	if order.DriverId != uint64(in.DriverId) {
		return nil, ErrDriverNotMatched
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusWaitPay,
		OperatorType: constants.OperatorDriver,
		OperatorId:   uint64(in.DriverId),
		Remark:       fmt.Sprintf("行程结束，实际距离=%dm，实际时长=%ds，实际费用=%d分", in.ActualDistanceM, in.ActualDurationS, in.ActualPriceCents),
	}
	ok, err := l.svcCtx.OrderRepository.FinishTrip(l.ctx, order.Id, uint64(in.DriverId), statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}

	return &proto.FinishTripResponse{
		OrderId:            in.OrderId,
		Status:             proto.OrderStatus_ORDER_STATUS_WAIT_PAY,
		PayableAmountCents: in.ActualPriceCents,
	}, nil
}

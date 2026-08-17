package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/keyutil"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCreateOrderLogic 创建订单逻辑对象。
func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateOrder 校验参数并创建待接单订单，同时写入创建状态日志。
func (l *CreateOrderLogic) CreateOrder(in *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	if err := validateCreateOrder(in); err != nil {
		return nil, err
	}

	order := &model.RideOrder{
		OrderNo:            keyutil.GenOrderID(),
		UserId:             uint64(in.UserId),
		DriverId:           0,
		CarType:            int8(in.CarType),
		FromAddress:        strings.TrimSpace(in.FromAddress),
		FromLongitude:      in.FromLongitude,
		FromLatitude:       in.FromLatitude,
		ToAddress:          strings.TrimSpace(in.ToAddress),
		ToLongitude:        in.ToLongitude,
		ToLatitude:         in.ToLatitude,
		EstimatedDistanceM: int(in.EstimatedDistanceM),
		EstimatedDurationS: int(in.EstimatedDurationS),
		EstimatedPrice:     float64(in.EstimatedPriceCents) / 100,
		Status:             constants.OrderStatusWaitAccept,
	}
	statusLog := &model.OrderStatusLog{
		FromStatus:   0,
		ToStatus:     constants.OrderStatusWaitAccept,
		OperatorType: constants.OperatorUser,
		OperatorId:   uint64(in.UserId),
		Remark:       "创建订单",
	}
	if err := l.svcCtx.OrderRepository.Create(l.ctx, order, statusLog); err != nil {
		return nil, err
	}
	if l.svcCtx.DispatchClient != nil {
		if _, err := l.svcCtx.DispatchClient.DispatchOrder(l.ctx, &dispatch.DispatchOrderRequest{
			OrderId:       int64(order.Id),
			FromLongitude: in.FromLongitude,
			FromLatitude:  in.FromLatitude,
			CarType:       in.CarType,
		}); err != nil {
			// 派单失败不阻塞下单，记录日志由后续任务补偿。
			l.Logger.Errorf("dispatch order %d failed: %v", order.Id, err)
		}
	}

	return &proto.CreateOrderResponse{
		OrderId:             int64(order.Id),
		OrderNo:             order.OrderNo,
		EstimatedPriceCents: in.EstimatedPriceCents,
		Status:              proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		CreatedAt:           order.CreatedAt.Unix(),
	}, nil
}

// validateCreateOrder 校验创建订单入参。
func validateCreateOrder(in *proto.CreateOrderRequest) error {
	if in.UserId <= 0 ||
		in.CarType < 1 || in.CarType > 3 ||
		strings.TrimSpace(in.FromAddress) == "" ||
		strings.TrimSpace(in.ToAddress) == "" ||
		in.FromLongitude < -180 || in.FromLongitude > 180 ||
		in.FromLatitude < -90 || in.FromLatitude > 90 ||
		in.ToLongitude < -180 || in.ToLongitude > 180 ||
		in.ToLatitude < -90 || in.ToLatitude > 90 ||
		in.EstimatedDistanceM < 0 ||
		in.EstimatedDurationS < 0 ||
		in.EstimatedPriceCents < 0 {
		return ErrInvalidOrderParams
	}
	return nil
}

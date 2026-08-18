package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// AcceptOrder 加 Redis 分布式锁保证同一订单只有一个司机接单成功。
func (l *AcceptOrderLogic) AcceptOrder(in *proto.AcceptOrderRequest) (*proto.AcceptOrderResponse, error) {
	if in.OrderId <= 0 || in.DriverId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	if l.svcCtx.Redis != nil {
		lockKey := fmt.Sprintf(constants.RedisOrderLock, in.OrderId)
		ok, err := l.svcCtx.Redis.SetNX(l.ctx, lockKey, "1", 10*time.Second).Result()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrOrderStatusNotAllowed
		}
		defer func() { _ = l.svcCtx.Redis.Del(l.ctx, lockKey).Err() }()
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

	// 闭环派单记录：司机接单后派单记录置为已接受。
	if err := l.svcCtx.OrderRepository.MarkDispatchAccepted(l.ctx, order.Id, uint64(in.DriverId)); err != nil {
		l.Logger.Errorf("mark dispatch accepted failed, orderId=%d driverId=%d: %v", order.Id, in.DriverId, err)
	}

	if l.svcCtx.EventBus != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"order_id":    in.OrderId,
			"driver_id":   in.DriverId,
			"from_status": constants.OrderStatusWaitAccept,
			"to_status":   constants.OrderStatusAccepted,
		})
		if err := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderStatusChanged, payload); err != nil {
			l.Logger.Errorf("publish order.status.changed failed: %v", err)
		}
	}

	return &proto.AcceptOrderResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_ACCEPTED,
	}, nil
}

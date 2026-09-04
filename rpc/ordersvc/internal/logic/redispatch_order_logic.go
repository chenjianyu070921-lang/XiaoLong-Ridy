package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"

	"github.com/zeromicro/go-zero/core/logx"
)

type RedispatchOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	Logger logx.Logger
}

func NewRedispatchOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedispatchOrderLogic {
	return &RedispatchOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RedispatchOrder 人工改派：
// 1) 校验订单处于待接单/已接单，解除司机绑定并回到待接单（写 admin 状态日志）；
// 2) 若指定 new_driver_id，直接绑定新司机为已接单，跳过自动派单；
// 3) 重发 order.created 事件（携带 exclude_driver_ids=原司机），由派单消费者重新派单。
func (l *RedispatchOrderLogic) RedispatchOrder(in *proto.RedispatchOrderRequest) (*proto.RedispatchOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	// 先读取订单当前状态与司机，用于校验与事件排除。
	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, repository.ErrOrderNotFound
	}
	prevStatus := order.Status
	prevDriver := order.DriverId

	// 允许改派的前置状态：待接单 / 已接单（司机未开始行程时可改派）。
	allowStatuses := []int8{
		constants.OrderStatusWaitAccept,
		constants.OrderStatusAccepted,
	}
	allow := false
	for _, s := range allowStatuses {
		if prevStatus == s {
			allow = true
			break
		}
	}
	if !allow {
		return nil, ErrInvalidOrderStatus
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   prevStatus,
		ToStatus:     constants.OrderStatusWaitAccept,
		OperatorType: constants.OperatorAdmin,
		OperatorId:   uint64(in.OperatorId),
		Remark:       "人工改派：" + strings.TrimSpace(in.Reason),
		CreatedAt:    time.Now(),
	}

	finalDriver, ok, err := l.svcCtx.OrderRepository.Redispatch(
		l.ctx, uint64(in.OrderId), uint64(in.NewDriverId), allowStatuses, statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidOrderStatus
	}

	// 重发 order.created 事件，排除原司机（若指定了新司机则不再自动派单，但仍发事件保持链路一致）。
	exclude := make([]int64, 0, 1)
	if prevDriver > 0 {
		exclude = append(exclude, int64(prevDriver))
	}
	if l.svcCtx.EventBus != nil {
		payload, _ := json.Marshal(orderCreatedEvent{
			OrderId:          in.OrderId,
			OrderNo:          order.OrderNo,
			FromLongitude:    order.FromLongitude,
			FromLatitude:     order.FromLatitude,
			CarType:          int32(order.CarType),
			CityCode:         order.CityCode,
			ExcludeDriverIds: exclude,
		})
		if pubErr := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderCreated, payload); pubErr != nil {
			l.Logger.Errorf("publish order.created for redispatch failed: %v", pubErr)
		}
	}

	// 事件不可用时回退为同步直派（排除原司机）。
	if l.svcCtx.DispatchClient != nil {
		if _, dErr := l.svcCtx.DispatchClient.DispatchOrder(l.ctx, &dispatch.DispatchOrderRequest{
			OrderId:          in.OrderId,
			FromLongitude:    order.FromLongitude,
			FromLatitude:     order.FromLatitude,
			CarType:          int32(order.CarType),
			CityCode:         order.CityCode,
			ExcludeDriverIds: exclude,
		}); dErr != nil {
			l.Logger.Errorf("sync dispatch on redispatch failed: %v", dErr)
		}
	}

	return &proto.RedispatchOrderResponse{
		OrderId:  in.OrderId,
		Status:   constants.OrderStatusWaitAccept,
		DriverId: int64(finalDriver),
	}, nil
}

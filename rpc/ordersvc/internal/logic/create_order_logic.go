package logic

import (
	"context"
	"encoding/json"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/keyutil"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

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

type orderCreatedEvent struct {
	OrderId       int64   `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
}

// CreateOrder 校验参数并创建待接单订单，同时写入创建状态日志并发布订单创建事件。
func (l *CreateOrderLogic) CreateOrder(in *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	if err := validateCreateOrder(in); err != nil {
		return nil, err
	}

	// 服务端计价快照：优先调 pricesvc 复核，失败降级为入参预估价格。
	estimatedPriceCents := l.estimatePriceSnapshot(in)

	order := &model.RideOrder{
		OrderNo:            keyutil.GenOrderID(),
		UserId:             uint64(in.UserId),
		DriverId:           0,
		CarType:            int8(in.CarType),
		CityCode:           strings.TrimSpace(in.CityCode),
		FromAddress:        strings.TrimSpace(in.FromAddress),
		FromLongitude:      in.FromLongitude,
		FromLatitude:       in.FromLatitude,
		ToAddress:          strings.TrimSpace(in.ToAddress),
		ToLongitude:        in.ToLongitude,
		ToLatitude:         in.ToLatitude,
		EstimatedDistanceM: int(in.EstimatedDistanceM),
		EstimatedDurationS: int(in.EstimatedDurationS),
		EstimatedPrice:     float64(estimatedPriceCents) / 100,
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

	// 优先发布 orderclient.created 事件，由 orderclient-event-consumer 触发派单。
	published := false
	if l.svcCtx.EventBus != nil {
		payload, _ := json.Marshal(orderCreatedEvent{
			OrderId:       int64(order.Id),
			OrderNo:       order.OrderNo,
			FromLongitude: in.FromLongitude,
			FromLatitude:  in.FromLatitude,
			CarType:       in.CarType,
			CityCode:      strings.TrimSpace(in.CityCode),
		})
		if err := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderCreated, payload); err != nil {
			l.Logger.Errorf("publish orderclient.created failed: %v", err)
		} else {
			published = true
		}
	}
	// 事件不可用时回退为同步直派，保证 demo 可跑通。
	if !published && l.svcCtx.DispatchClient != nil {
		if _, err := l.svcCtx.DispatchClient.DispatchOrder(l.ctx, &dispatch.DispatchOrderRequest{
			OrderId:       int64(order.Id),
			FromLongitude: in.FromLongitude,
			FromLatitude:  in.FromLatitude,
			CarType:       in.CarType,
			CityCode:      in.CityCode,
		}); err != nil {
			l.Logger.Errorf("dispatch orderclient %d failed: %v", order.Id, err)
		}
	}

	return &proto.CreateOrderResponse{
		OrderId:             int64(order.Id),
		OrderNo:             order.OrderNo,
		EstimatedPriceCents: estimatedPriceCents,
		Status:              proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		CreatedAt:           order.CreatedAt.Unix(),
	}, nil
}

// defaultCityCode 兜底城市编码，与 api 网关默认值保持一致。
const defaultCityCode = "110000"

// estimatePriceSnapshot 调 pricesvc 落计价快照；客户端缺失或计价失败时降级为入参预估价格。
func (l *CreateOrderLogic) estimatePriceSnapshot(in *proto.CreateOrderRequest) int64 {
	if l.svcCtx.PriceClient == nil {
		return in.EstimatedPriceCents
	}
	cityCode := strings.TrimSpace(in.CityCode)
	if cityCode == "" {
		cityCode = defaultCityCode
	}
	resp, err := l.svcCtx.PriceClient.EstimatePrice(l.ctx, &price.EstimatePriceRequest{
		UserId:    in.UserId,
		CityCode:  cityCode,
		CarType:   in.CarType,
		DistanceM: in.EstimatedDistanceM,
		DurationS: in.EstimatedDurationS,
	})
	if err != nil || resp == nil || resp.TotalCents <= 0 {
		l.Logger.Errorf("estimate price failed, fallback to input %d: %v", in.EstimatedPriceCents, err)
		return in.EstimatedPriceCents
	}
	return resp.TotalCents
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

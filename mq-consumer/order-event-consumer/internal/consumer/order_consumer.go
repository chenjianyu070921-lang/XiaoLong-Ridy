package consumer

import (
	"context"
	"encoding/json"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
)

// OrderCreatedEvent 与 ordersvc 发布的事件字段保持一致。
type OrderCreatedEvent struct {
	OrderId       int64   `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
}

// OrderConsumer 消费订单事件并触发派单。
type OrderConsumer struct {
	svcCtx *svc.ServiceContext
}

func NewOrderConsumer(svcCtx *svc.ServiceContext) *OrderConsumer {
	return &OrderConsumer{svcCtx: svcCtx}
}

// Start 阻塞消费订单事件流。
func (c *OrderConsumer) Start(ctx context.Context) error {
	return c.svcCtx.EventBus.Consume(ctx, "order-event-consumer", c.handle)
}

func (c *OrderConsumer) handle(ctx context.Context, topic string, payload []byte) error {
	if topic != constants.TopicOrderCreated {
		return nil
	}
	var evt OrderCreatedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return err
	}
	_, err := c.svcCtx.DispatchClient.DispatchOrder(ctx, &dispatch.DispatchOrderRequest{
		OrderId:       evt.OrderId,
		FromLongitude: evt.FromLongitude,
		FromLatitude:  evt.FromLatitude,
		CarType:       evt.CarType,
		CityCode:      evt.CityCode,
	})
	return err
}

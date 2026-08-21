package consumer

import (
	"context"
	"encoding/json"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
)

const topicOrderPaid = "order.paid"

// OrderCreatedEvent 与 ordersvc 发布的事件字段保持一致。
type OrderCreatedEvent struct {
	OrderId       int64   `json:"order_id"`
	OrderNo       string  `json:"order_no"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
}

// OrderPaidEvent 对齐 paysvc 支付成功后发布的 order.paid 事件载荷。
type OrderPaidEvent struct {
	OrderID     int64  `json:"order_id"`
	PaymentNo   string `json:"payment_no"`
	AmountCents int64  `json:"amount_cents"`
	PaidAt      int64  `json:"paid_at"`
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
	return c.svcCtx.EventBus.Consume(ctx, "orderclient-event-consumer", c.handle)
}

func (c *OrderConsumer) handle(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case constants.TopicOrderCreated:
		return c.handleOrderCreated(ctx, payload)
	case constants.TopicDispatchNew:
		return c.handleDispatchNew(ctx, payload)
	case topicOrderPaid:
		return c.handleOrderPaid(ctx, payload)
	default:
		return nil
	}
}

func (c *OrderConsumer) handleOrderCreated(ctx context.Context, payload []byte) error {
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

func (c *OrderConsumer) handleOrderPaid(ctx context.Context, payload []byte) error {
	var evt OrderPaidEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return err
	}
	_, err := c.svcCtx.OrderClient.ConfirmPaid(ctx, &order.ConfirmPaidRequest{
		OrderId:     evt.OrderID,
		PaymentNo:   evt.PaymentNo,
		AmountCents: evt.AmountCents,
		PaidAt:      evt.PaidAt,
	})
	return err
}

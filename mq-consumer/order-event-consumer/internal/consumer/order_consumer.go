package consumer

import (
	"context"
	"encoding/json"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
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

// Start 阻塞消费订单事件流（P1-M4-11）。
// 与支付模块对齐改用 Kafka：单消费者组订阅全部事件 topic，
// 由 Kafka 消费组机制保证消息不重复投递（partition 级别 exactly-once-in-order），
// 在 handler 内按 topic 分发到对应处理函数，避免拆多个消费组互相抢分区。
func (c *OrderConsumer) Start(ctx context.Context) error {
	const group = "orderclient-event-consumer"
	return c.svcCtx.EventBus.Consume(ctx, group, c.dispatchHandler,
		constants.TopicOrderCreated, constants.TopicDispatchNew, constants.TopicOrderPaid)
}

// dispatchHandler 按 topic 分发到对应事件处理函数；未知 topic 直接忽略（不阻塞消费）。
func (c *OrderConsumer) dispatchHandler(ctx context.Context, topic string, payload []byte) error {
	switch topic {
	case constants.TopicOrderCreated:
		return c.handleOrderCreated(ctx, payload)
	case constants.TopicDispatchNew:
		return c.handleDispatchNew(ctx, payload)
	case constants.TopicOrderPaid:
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

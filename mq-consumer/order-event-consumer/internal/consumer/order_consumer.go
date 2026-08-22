package consumer

import (
	"context"
	"encoding/json"
	"sync"

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
// 按主题拆分为 3 个并行消费 goroutine（同一消费者组，Redis 消费组保证消息不重复投递；
// 消费端 Consume 内已按 topics 过滤，非本主题消息直接 ACK 跳过），
// 避免单个事件慢处理（如派单 RPC 阻塞）拖慢其他主题的消费吞吐。
func (c *OrderConsumer) Start(ctx context.Context) error {
	const group = "orderclient-event-consumer"
	var wg sync.WaitGroup
	errCh := make(chan error, 3)
	wg.Add(3)
	go func() {
		defer wg.Done()
		errCh <- c.svcCtx.EventBus.Consume(ctx, group, func(ctx context.Context, _ string, payload []byte) error {
			return c.handleOrderCreated(ctx, payload)
		}, constants.TopicOrderCreated)
	}()
	go func() {
		defer wg.Done()
		errCh <- c.svcCtx.EventBus.Consume(ctx, group, func(ctx context.Context, _ string, payload []byte) error {
			return c.handleDispatchNew(ctx, payload)
		}, constants.TopicDispatchNew)
	}()
	go func() {
		defer wg.Done()
		errCh <- c.svcCtx.EventBus.Consume(ctx, group, func(ctx context.Context, _ string, payload []byte) error {
			return c.handleOrderPaid(ctx, payload)
		}, constants.TopicOrderPaid)
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
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

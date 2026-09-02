package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"

	"github.com/zeromicro/go-zero/core/logx"
)

// OrderCreatedEvent 与 ordersvc 发布的事件字段保持一致。
type OrderCreatedEvent struct {
	OrderId          int64   `json:"order_id"`
	OrderNo          string  `json:"order_no"`
	FromLongitude    float64 `json:"from_longitude"`
	FromLatitude     float64 `json:"from_latitude"`
	CarType          int32   `json:"car_type"`
	CityCode         string  `json:"city_code"`
	ExcludeDriverIds []int64 `json:"exclude_driver_ids"` // 改派/重派时排除的司机
}

// OrderRefundedEvent 与 ordersvc 发布的 order.refunded 事件字段保持一致。
type OrderRefundedEvent struct {
	OrderId      int64  `json:"order_id"`
	OrderNo      string `json:"order_no"`
	RefundNo     string `json:"refund_no"`
	RefundCents  int64  `json:"refund_cents"`
	OperatorId   int64  `json:"operator_id"`
	OperatorType string `json:"operator_type"`
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
		constants.TopicOrderCreated, constants.TopicDispatchNew, constants.TopicOrderPaid, constants.TopicOrderRefunded)
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
	case constants.TopicOrderRefunded:
		return c.handleOrderRefunded(ctx, payload)
	default:
		return nil
	}
}

func (c *OrderConsumer) handleOrderCreated(ctx context.Context, payload []byte) error {
	var evt OrderCreatedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		logx.WithContext(ctx).Errorf("order.created consume decode failed: payloadBytes=%d error=%v", len(payload), err)
		return err
	}
	logx.WithContext(ctx).Infof("order.created consumed: orderId=%d orderNo=%s cityCode=%s carType=%d fromLng=%.6f fromLat=%.6f payloadBytes=%d",
		evt.OrderId, evt.OrderNo, evt.CityCode, evt.CarType, evt.FromLongitude, evt.FromLatitude, len(payload))
	resp, err := c.svcCtx.DispatchClient.DispatchOrder(ctx, &dispatch.DispatchOrderRequest{
		OrderId:          evt.OrderId,
		FromLongitude:    evt.FromLongitude,
		FromLatitude:     evt.FromLatitude,
		CarType:          evt.CarType,
		CityCode:         evt.CityCode,
		ExcludeDriverIds: evt.ExcludeDriverIds,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("order.created dispatch rpc failed: orderId=%d orderNo=%s cityCode=%s error=%v",
			evt.OrderId, evt.OrderNo, evt.CityCode, err)
		return err
	}
	logx.WithContext(ctx).Infof("order.created dispatch rpc succeeded: orderId=%d orderNo=%s cityCode=%s responseOrderId=%d dispatchRecords=%d",
		evt.OrderId, evt.OrderNo, evt.CityCode, resp.GetOrderId(), len(resp.GetList()))
	return err
}

// handleOrderRefunded 处理订单退款事件，完成支付通道退款。
// 处理失败必须返回错误，由 Kafka 消费器写入退款主题 DLQ，避免订单已退款但资金未回退
// 时被静默吞掉。
func (c *OrderConsumer) handleOrderRefunded(ctx context.Context, payload []byte) error {
	var evt OrderRefundedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return err
	}
	if c.svcCtx.PayClient == nil {
		return fmt.Errorf("pay client not configured")
	}
	payment, err := c.svcCtx.PayClient.GetPayment(ctx, &pay.GetPaymentRequest{OrderId: evt.OrderId})
	if err != nil {
		return fmt.Errorf("get payment by order %d: %w", evt.OrderId, err)
	}
	if payment == nil || payment.PaymentNo == "" {
		return fmt.Errorf("payment not found for order %d", evt.OrderId)
	}
	resp, err := c.svcCtx.PayClient.RefundPayment(ctx, &pay.RefundPaymentRequest{
		PaymentNo:         payment.PaymentNo,
		RefundAmountCents: evt.RefundCents,
		Reason:            "订单退款：" + evt.RefundNo,
		RefundNo:          evt.RefundNo,
	})
	if err != nil {
		return fmt.Errorf("refund payment %s: %w", payment.PaymentNo, err)
	}
	if resp == nil || !resp.Success {
		return fmt.Errorf("refund payment %s unsuccessful", payment.PaymentNo)
	}
	logx.Infof("order.refunded processed: order=%d refundNo=%s paymentNo=%s cents=%d operator=%s(%d)",
		evt.OrderId, evt.RefundNo, payment.PaymentNo, evt.RefundCents, evt.OperatorType, evt.OperatorId)
	return nil
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

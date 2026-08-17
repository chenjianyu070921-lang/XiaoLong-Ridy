// Package mq 提供 Kafka 生产者封装与事件消息定义。
package mq

import "encoding/json"

// OrderPaidEvent 订单支付成功事件消息体。
type OrderPaidEvent struct {
	OrderId     int64  `json:"order_id"`     // 订单ID
	PaymentNo   string `json:"payment_no"`   // 平台支付单号
	AmountCents int64  `json:"amount_cents"` // 支付金额（分）
	PaidAt      int64  `json:"paid_at"`      // 支付时间（Unix 秒）
}

// EncodeOrderPaidEvent 将支付成功事件序列化为 JSON 字节。
func EncodeOrderPaidEvent(e *OrderPaidEvent) ([]byte, error) {
	return json.Marshal(e)
}

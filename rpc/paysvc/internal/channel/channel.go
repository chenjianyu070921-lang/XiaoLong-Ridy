// Package channel 定义支付渠道抽象，隔离第三方支付差异。
// 本期提供 MockChannel 实现，后续可替换为微信/支付宝真实渠道。
package channel

import "context"

// 支付渠道标识
const (
	Wechat  = "wechat"
	Alipay  = "alipay"
	Balance = "balance"
)

// OrderResult 第三方下单结果。
type OrderResult struct {
	TransactionId string // 第三方支付流水号
	PayParams     string // 支付参数（预支付 ID / 二维码串等）
}

// RefundResult 第三方退款结果。
type RefundResult struct {
	RefundNo string // 第三方退款单号
}

// PayChannel 支付渠道接口。
type PayChannel interface {
	// Name 返回渠道标识。
	Name() string
	// CreateOrder 预下单：向第三方发起统一下单。
	CreateOrder(ctx context.Context, paymentNo string, amountCents int64) (*OrderResult, error)
	// Refund 退款：向第三方发起退款。
	Refund(ctx context.Context, paymentNo string, refundNo string, amountCents int64) (*RefundResult, error)
}

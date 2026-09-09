package channel

import (
	"context"
	"fmt"
)

// MockChannel 模拟支付渠道：返回模拟流水号与支付参数。
type MockChannel struct {
	name string
}

// NewMockChannel 创建模拟渠道。
func NewMockChannel(name string) *MockChannel {
	return &MockChannel{name: name}
}

func (m *MockChannel) Name() string {
	return m.name
}

// CreateOrder 模拟下单：生成假的 transaction_id 与 pay_params。
func (m *MockChannel) CreateOrder(ctx context.Context, paymentNo string, amountCents int64) (*OrderResult, error) {
	// 模拟第三方流水号：渠道前缀 + 支付单号
	transactionId := fmt.Sprintf("%s_%s", m.name, paymentNo)
	// 模拟支付参数（如预支付 ID）
	payParams := fmt.Sprintf("mock://%s?payment_no=%s&amount=%d", m.name, paymentNo, amountCents)

	return &OrderResult{
		TransactionId: transactionId,
		PayParams:     payParams,
	}, nil
}

// Refund 模拟退款：生成假的退款单号。
func (m *MockChannel) Refund(ctx context.Context, paymentNo string, refundNo string, amountCents int64) (*RefundResult, error) {
	return &RefundResult{
		RefundNo: refundNo,
	}, nil
}

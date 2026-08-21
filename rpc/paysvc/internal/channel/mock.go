package channel

import (
	"context"
	"fmt"
)

// ────── M5-9 真实渠道接入现状 ──────
//
// 本文件实现的是支付渠道的 Mock 占位实现，仅用于：
//   1) 单元测试与本地调试：不需要沙箱账号也能跑通 CreatePayment/RefundPayment 业务流程；
//   2) 沙箱密钥缺失时的降级路径：当 paysvc.yaml / 环境变量未配置沙箱密钥时，
//      service_context.newAlipayChannel 会返回 nil，GetChannel 走 MockChannel 兜底。
//
// 生产环境真实接入路径（参考接入步骤，不在本期交付范围内）：
//   1. 替换 MockChannel.NewMockChannel 调用为真实渠道构造器：
//        wechat:  wechat.NewWechatChannel(cfg)
//        alipay:  channel.NewAlipayChannel(alipayCfg) ← SDK 已封装在 channel/alipay.go
//        balance: 余额走账户服务扣减，无需第三方 SDK
//   2. 将本次提交引入的 MockChannel 注释为"internal/test only"，仅在 _test.go 文件中可见。
//   3. 在 ChannelSelector（即 GetChannel）中按渠道名作类型分派，按配置启用真实渠道。
//
// 当下在本服务目录内已提供：
//   - channel/alipay.go: 基于 smartwalle/alipay SDK 的 AlipayChannel（CreateOrder/Refund 全实现）。
//   - channel/verify.go: AlipayVerifier（RSA2 验签）。
//
// 因此支付宝渠道在密钥齐全时即走真实路径，本期第三方下单/退款的实现度仅取决于密钥是否可用。

// MockChannel 模拟支付渠道：返回模拟流水号与支付参数。
// ⚠️ 注意：此实现不应进入生产路径，仅用于单元测试与本地降级。
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

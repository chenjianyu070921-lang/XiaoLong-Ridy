package channel

import (
	"context"
	"net/url"

	"github.com/smartwalle/alipay/v3"
)

// SignVerifier 支付回调签名验签接口，便于业务层解耦与测试替换。
type SignVerifier interface {
	// Verify 校验回调签名的合法性，返回 nil 表示验签通过。
	Verify(ctx context.Context, notifyRaw string) error
}

// AlipayVerifier 基于支付宝官方 SDK 的验签实现（RSA2）。
type AlipayVerifier struct {
	client *alipay.Client
}

// NewAlipayVerifier 创建支付宝验签器。
//
// appId           支付宝应用 ID
// privateKey      应用私钥
// alipayPublicKey 支付宝公钥
// sandbox         是否沙箱环境
func NewAlipayVerifier(appId, privateKey, alipayPublicKey string, sandbox bool) (*AlipayVerifier, error) {
	client, err := alipay.New(appId, privateKey, !sandbox)
	if err != nil {
		return nil, err
	}
	if err := client.LoadAliPayPublicKey(alipayPublicKey); err != nil {
		return nil, err
	}
	return &AlipayVerifier{client: client}, nil
}

// Verify 解析原始回调参数并验签。
func (v *AlipayVerifier) Verify(ctx context.Context, notifyRaw string) error {
	values, err := url.ParseQuery(notifyRaw)
	if err != nil {
		return err
	}
	return v.client.VerifySign(ctx, values)
}

// MockVerifier 模拟验签器，用于开发与测试。
// Err 为 nil 表示验签通过，否则返回该错误。
type MockVerifier struct {
	Err error
}

// Verify 按配置返回验签结果。
func (m *MockVerifier) Verify(ctx context.Context, notifyRaw string) error {
	return m.Err
}

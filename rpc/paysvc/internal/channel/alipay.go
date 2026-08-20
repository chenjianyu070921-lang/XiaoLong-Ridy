package channel

import (
	"context"
	"fmt"
	"strconv"

	commonalipay "XiaoLong-Ridy/common/alipay"
	"XiaoLong-Ridy/common/priceutil"

	"github.com/smartwalle/alipay/v3"
)

// AlipayChannel 支付宝手机网站支付渠道（alipay.trade.wap.pay）。
// 负责预下单生成支付跳转链接，以及退款。
type AlipayChannel struct {
	client    *alipay.Client
	notifyURL string
	returnURL string

	// 可注入的下单/退款执行函数，便于测试替换底层 SDK 调用。
	createOrderFn func(ctx context.Context, param alipay.TradeWapPay) (string, error)
	refundFn      func(ctx context.Context, param alipay.TradeRefund) (*RefundResult, error)
}

// NewAlipayChannel 根据配置创建支付宝渠道。
//
// cfg.Sandbox 为 true 时走沙箱环境网关（alipay.New 第三参传 false）。
func NewAlipayChannel(cfg commonalipay.Config) (*AlipayChannel, error) {
	c := cfg.WithDefaults()
	client, err := alipay.New(c.AppId, c.PrivateKey, !c.Sandbox)
	if err != nil {
		return nil, fmt.Errorf("init alipay client: %w", err)
	}
	if err := client.LoadAliPayPublicKey(c.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("load alipay public key: %w", err)
	}
	ch := &AlipayChannel{
		client:    client,
		notifyURL: c.NotifyUrl,
		returnURL: c.ReturnUrl,
	}
	ch.createOrderFn = ch.doCreateOrder
	ch.refundFn = ch.doRefund
	return ch, nil
}

// NewAlipayChannelWithClient 注入已构造的 SDK client，便于测试。
func NewAlipayChannelWithClient(client *alipay.Client, notifyURL, returnURL string) *AlipayChannel {
	ch := &AlipayChannel{client: client, notifyURL: notifyURL, returnURL: returnURL}
	ch.createOrderFn = ch.doCreateOrder
	ch.refundFn = ch.doRefund
	return ch
}

func (c *AlipayChannel) Name() string {
	return Alipay
}

// CreateOrder 手机网站支付预下单：生成支付跳转链接（url.URL）作为 pay_params。
func (c *AlipayChannel) CreateOrder(ctx context.Context, paymentNo string, amountCents int64) (*OrderResult, error) {
	param := alipay.TradeWapPay{
		Trade: alipay.Trade{
			Subject:     "打车行程支付",
			OutTradeNo:  paymentNo,
			TotalAmount: yuanStr(amountCents),
			ProductCode: "QUICK_WAP_WAY",
			NotifyURL:   c.notifyURL,
			ReturnURL:   c.returnURL,
		},
	}
	payParams, err := c.createOrderFn(ctx, param)
	if err != nil {
		return nil, err
	}
	// 预下单阶段尚未拿到支付宝流水号，transaction_id 留空，待回调回填。
	return &OrderResult{TransactionId: "", PayParams: payParams}, nil
}

// Refund 调用支付宝退款，返回退款单号。
func (c *AlipayChannel) Refund(ctx context.Context, paymentNo string, refundNo string, amountCents int64) (*RefundResult, error) {
	return c.refundFn(ctx, alipay.TradeRefund{
		OutTradeNo:   paymentNo,
		RefundAmount: yuanStr(amountCents),
		OutRequestNo: refundNo,
		RefundReason: "乘客申请退款",
	})
}

// doCreateOrder 调用支付宝 SDK 生成支付链接。
func (c *AlipayChannel) doCreateOrder(ctx context.Context, param alipay.TradeWapPay) (string, error) {
	payURL, err := c.client.TradeWapPay(param)
	if err != nil {
		return "", fmt.Errorf("alipay wap pay: %w", err)
	}
	return payURL.String(), nil
}

// doRefund 调用支付宝 SDK 退款。
// 退款单号（out_request_no）由平台生成并在入参传入；SDK 成功响应不回传独立退款单号。
func (c *AlipayChannel) doRefund(ctx context.Context, param alipay.TradeRefund) (*RefundResult, error) {
	rsp, err := c.client.TradeRefund(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("alipay refund: %w", err)
	}
	if rsp.Code != "10000" {
		return nil, fmt.Errorf("alipay refund failed: %s %s", rsp.Code, rsp.SubMsg)
	}
	// 以平台生成的退款请求号作为退款单号。
	return &RefundResult{RefundNo: param.OutRequestNo}, nil
}

// yuanStr 将「分」转为支付宝要求的元字符串（精确到两位小数），如 2950 -> "29.50"。
func yuanStr(cents int64) string {
	return strconv.FormatFloat(priceutil.CentsToYuan(cents), 'f', 2, 64)
}

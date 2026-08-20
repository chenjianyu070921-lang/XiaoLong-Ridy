package channel

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/smartwalle/alipay/v3"
)

// stubAlipayChannel 构造一个注入了桩函数、无需真实密钥/网络的 AlipayChannel。
func stubAlipayChannel(t *testing.T, createErr error, refundCode string, refundNo string) *AlipayChannel {
	t.Helper()
	ch := &AlipayChannel{notifyURL: "http://notify.example/pay", returnURL: "http://return.example/result"}
	ch.createOrderFn = func(ctx context.Context, p alipay.TradeWapPay) (string, error) {
		if createErr != nil {
			return "", createErr
		}
		return "https://openapi.alipay.com/gateway.do?out_trade_no=" + p.OutTradeNo, nil
	}
	ch.refundFn = func(ctx context.Context, p alipay.TradeRefund) (*RefundResult, error) {
		if refundCode != "10000" {
			return nil, errors.New("alipay refund failed: " + refundCode)
		}
		return &RefundResult{RefundNo: refundNo}, nil
	}
	return ch
}

func TestAlipayChannel_Name(t *testing.T) {
	if got := stubAlipayChannel(t, nil, "", "").Name(); got != "alipay" {
		t.Fatalf("Name() = %q, want alipay", got)
	}
}

func TestAlipayChannel_CreateOrder(t *testing.T) {
	ch := stubAlipayChannel(t, nil, "", "")
	res, err := ch.CreateOrder(context.Background(), "PAY20260820100000001", 2950)
	if err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}
	if res.TransactionId != "" {
		t.Fatalf("预下单阶段 TransactionId 应为空, got %q", res.TransactionId)
	}
	if !strings.Contains(res.PayParams, "out_trade_no=PAY20260820100000001") {
		t.Fatalf("pay_params 应包含支付单号, got %q", res.PayParams)
	}
}

func TestAlipayChannel_CreateOrder_Error(t *testing.T) {
	ch := stubAlipayChannel(t, errors.New("boom"), "", "")
	if _, err := ch.CreateOrder(context.Background(), "PAY1", 100); err == nil {
		t.Fatal("期望下单失败返回错误，实际为 nil")
	}
}

func TestAlipayChannel_Refund_OK(t *testing.T) {
	ch := stubAlipayChannel(t, nil, "10000", "RF20260820001")
	res, err := ch.Refund(context.Background(), "PAY1", "RF20260820001", 500)
	if err != nil {
		t.Fatalf("Refund error: %v", err)
	}
	if res.RefundNo != "RF20260820001" {
		t.Fatalf("RefundNo = %q", res.RefundNo)
	}
}

func TestAlipayChannel_Refund_Fail(t *testing.T) {
	ch := stubAlipayChannel(t, nil, "40004", "")
	if _, err := ch.Refund(context.Background(), "PAY1", "RF1", 500); err == nil {
		t.Fatal("期望退款失败返回错误，实际为 nil")
	}
}

func TestYuanStr(t *testing.T) {
	cases := []struct {
		cents int64
		want  string
	}{
		{2950, "29.50"},
		{100, "1.00"},
		{0, "0.00"},
		{999, "9.99"},
	}
	for _, c := range cases {
		if got := yuanStr(c.cents); got != c.want {
			t.Errorf("yuanStr(%d) = %q, want %q", c.cents, got, c.want)
		}
	}
}

package logic

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"google.golang.org/grpc"
)

// mockPaySvc 实现 payclient.Pay 接口，用于单测注入。
type mockPaySvc struct {
	notifyFn func(ctx context.Context, in *proto.NotifyPaymentRequest) (*proto.NotifyPaymentResponse, error)
}

func (m *mockPaySvc) CreatePayment(ctx context.Context, in *proto.CreatePaymentRequest, opts ...grpc.CallOption) (*proto.CreatePaymentResponse, error) {
	return nil, nil
}
func (m *mockPaySvc) NotifyPayment(ctx context.Context, in *proto.NotifyPaymentRequest, opts ...grpc.CallOption) (*proto.NotifyPaymentResponse, error) {
	return m.notifyFn(ctx, in)
}
func (m *mockPaySvc) GetPayment(ctx context.Context, in *proto.GetPaymentRequest, opts ...grpc.CallOption) (*proto.GetPaymentResponse, error) {
	return nil, nil
}
func (m *mockPaySvc) RefundPayment(ctx context.Context, in *proto.RefundPaymentRequest, opts ...grpc.CallOption) (*proto.RefundPaymentResponse, error) {
	return nil, nil
}
func (m *mockPaySvc) SettleOrder(ctx context.Context, in *proto.SettleOrderRequest, opts ...grpc.CallOption) (*proto.SettleOrderResponse, error) {
	return nil, nil
}

func TestPayCallbackLogic_HandleAlipayNotify_Success(t *testing.T) {
	paySvc := &mockPaySvc{notifyFn: func(ctx context.Context, in *proto.NotifyPaymentRequest) (*proto.NotifyPaymentResponse, error) {
		if in.PaymentNo != "PAY20260828120000000001" {
			t.Fatalf("unexpected payment_no: %s", in.PaymentNo)
		}
		if in.TotalAmountCents != 2950 {
			t.Fatalf("unexpected amount cents: %d", in.TotalAmountCents)
		}
		if in.NotifyRaw == "" {
			t.Fatalf("notify_raw must not be empty")
		}
		return &proto.NotifyPaymentResponse{Success: true, Message: "ok"}, nil
	}}

	ctx := &svc.ServiceContext{PaySvc: paySvc}
	l := NewPayCallbackLogic(ctx)

	form := url.Values{}
	form.Set("out_trade_no", "PAY20260828120000000001")
	form.Set("trade_status", "TRADE_SUCCESS")
	form.Set("trade_no", "2026082822001401640500000001")
	form.Set("total_amount", "29.50")
	form.Set("gmt_payment", "2026-08-28 12:00:00")
	req := httptest.NewRequest("POST", "/api/pay/callback/alipay", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := l.HandleAlipayNotify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "success" {
		t.Fatalf("want success, got %s", body)
	}
}

func TestPayCallbackLogic_HandleAlipayNotify_RPCFail(t *testing.T) {
	paySvc := &mockPaySvc{notifyFn: func(ctx context.Context, in *proto.NotifyPaymentRequest) (*proto.NotifyPaymentResponse, error) {
		return &proto.NotifyPaymentResponse{Success: false, Message: "verify fail"}, nil
	}}
	ctx := &svc.ServiceContext{PaySvc: paySvc}
	l := NewPayCallbackLogic(ctx)

	form := url.Values{}
	form.Set("out_trade_no", "PAY20260828120000000002")
	form.Set("trade_status", "TRADE_SUCCESS")
	form.Set("total_amount", "10.00")
	req := httptest.NewRequest("POST", "/api/pay/callback/alipay", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, _ := l.HandleAlipayNotify(req)
	if body != "fail" {
		t.Fatalf("want fail, got %s", body)
	}
}

func TestPayCallbackLogic_HandleAlipayNotify_MissingOutTradeNo(t *testing.T) {
	ctx := &svc.ServiceContext{}
	l := NewPayCallbackLogic(ctx)

	req := httptest.NewRequest("POST", "/api/pay/callback/alipay", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, _ := l.HandleAlipayNotify(req)
	if body != "fail" {
		t.Fatalf("want fail, got %s", body)
	}
}

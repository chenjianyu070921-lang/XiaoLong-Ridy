package handler

import (
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/proto"
)

func TestBuildNotifyRequest_AlipayStandard(t *testing.T) {
	raw := map[string]string{
		"out_trade_no": "PAY202308220001",
		"trade_no":     "2023082222001156789012345678",
		"total_amount": "12.34",
		"gmt_payment":  "2023-08-22 10:20:30",
	}

	req := buildNotifyRequest(raw)
	if req.PaymentNo != "PAY202308220001" {
		t.Errorf("PaymentNo = %s, want PAY202308220001", req.PaymentNo)
	}
	if req.TransactionId != "2023082222001156789012345678" {
		t.Errorf("TransactionId = %s, want 2023082222001156789012345678", req.TransactionId)
	}
	if req.TotalAmountCents != 1234 {
		t.Errorf("TotalAmountCents = %d, want 1234", req.TotalAmountCents)
	}
	wantPaidAt, _ := time.Parse("2006-01-02 15:04:05", raw["gmt_payment"])
	if req.PaidAt != wantPaidAt.Unix() {
		t.Errorf("PaidAt = %d, want %d", req.PaidAt, wantPaidAt.Unix())
	}
}

func TestBuildNotifyRequest_Custom(t *testing.T) {
	raw := map[string]string{
		"payment_no":    "PAY123",
		"transaction_id": "TX123",
		"totalAmount":   "56.78",
		"paid_at":       "1692669000",
	}

	req := buildNotifyRequest(raw)
	if req.PaymentNo != "PAY123" {
		t.Errorf("PaymentNo = %s, want PAY123", req.PaymentNo)
	}
	if req.TransactionId != "TX123" {
		t.Errorf("TransactionId = %s, want TX123", req.TransactionId)
	}
	if req.TotalAmountCents != 5678 {
		t.Errorf("TotalAmountCents = %d, want 5678", req.TotalAmountCents)
	}
	if req.PaidAt != 1692669000 {
		t.Errorf("PaidAt = %d, want 1692669000", req.PaidAt)
	}
}

func TestParseNotifyForm(t *testing.T) {
	body := "a=1&b=2&c=hello%20world"
	got := ParseNotifyForm(body)
	if got["a"] != "1" || got["b"] != "2" || got["c"] != "hello%20world" {
		t.Errorf("unexpected parse result: %v", got)
	}
}

func TestBuildNotifyRequest_ProtoType(t *testing.T) {
	// 保证 proto 类型仍兼容当前 handler 字段赋值（编译期检查）。
	_ = &proto.NotifyPaymentRequest{
		PaymentNo:        "PAY",
		TransactionId:    "TX",
		TotalAmountCents: 100,
		PaidAt:           0,
	}
}

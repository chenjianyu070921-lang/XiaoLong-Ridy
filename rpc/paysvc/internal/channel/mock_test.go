package channel

import (
	"context"
	"strings"
	"testing"
)

func TestMockChannel_CreateOrder(t *testing.T) {
	ch := NewMockChannel(Wechat)
	res, err := ch.CreateOrder(context.Background(), "PAY20260813120000001", 2500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.TransactionId, "wechat_") {
		t.Errorf("transaction id = %q, want prefix wechat_", res.TransactionId)
	}
	if res.TransactionId != "wechat_PAY20260813120000001" {
		t.Errorf("transaction id = %q, want wechat_PAY20260813120000001", res.TransactionId)
	}
	if res.PayParams == "" {
		t.Error("pay params should not be empty")
	}
}

func TestMockChannel_Name(t *testing.T) {
	ch := NewMockChannel(Alipay)
	if ch.Name() != "alipay" {
		t.Errorf("name = %q, want alipay", ch.Name())
	}
}

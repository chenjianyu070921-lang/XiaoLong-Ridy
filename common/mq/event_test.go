package mq

import (
	"encoding/json"
	"testing"
)

func TestEncodeOrderPaidEvent(t *testing.T) {
	e := &OrderPaidEvent{
		OrderId:     1001,
		PaymentNo:   "PAY202608140001",
		AmountCents: 2500,
		PaidAt:      1753065600,
	}
	b, err := EncodeOrderPaidEvent(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded OrderPaidEvent
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded.OrderId != 1001 {
		t.Errorf("order_id = %d, want 1001", decoded.OrderId)
	}
	if decoded.PaymentNo != "PAY202608140001" {
		t.Errorf("payment_no = %s, want PAY202608140001", decoded.PaymentNo)
	}
	if decoded.AmountCents != 2500 {
		t.Errorf("amount_cents = %d, want 2500", decoded.AmountCents)
	}
	if decoded.PaidAt != 1753065600 {
		t.Errorf("paid_at = %d, want 1753065600", decoded.PaidAt)
	}
}

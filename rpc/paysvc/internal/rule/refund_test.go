package rule

import (
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
)

func TestValidateRefund_Ok(t *testing.T) {
	// 已支付 10000 分，已退 0，本次退 3000 → 合法
	if err := ValidateRefund(model.PaymentStatusPaid, 10000, 0, 3000); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidateRefund_NotPaid(t *testing.T) {
	// 待支付状态不能退款
	err := ValidateRefund(model.PaymentStatusPending, 10000, 0, 3000)
	if !errors.Is(err, ErrRefundNotAllowed) {
		t.Errorf("expected ErrRefundNotAllowed, got %v", err)
	}
}

func TestValidateRefund_ZeroAmount(t *testing.T) {
	err := ValidateRefund(model.PaymentStatusPaid, 10000, 0, 0)
	if !errors.Is(err, ErrRefundAmountInvalid) {
		t.Errorf("expected ErrRefundAmountInvalid, got %v", err)
	}
}

func TestValidateRefund_NegativeAmount(t *testing.T) {
	err := ValidateRefund(model.PaymentStatusPaid, 10000, 0, -100)
	if !errors.Is(err, ErrRefundAmountInvalid) {
		t.Errorf("expected ErrRefundAmountInvalid, got %v", err)
	}
}

func TestValidateRefund_Exceed(t *testing.T) {
	// 已支付 10000，已退 8000，本次再退 3000 → 超 1000
	err := ValidateRefund(model.PaymentStatusPaid, 10000, 8000, 3000)
	if !errors.Is(err, ErrRefundExceed) {
		t.Errorf("expected ErrRefundExceed, got %v", err)
	}
}

func TestValidateRefund_PartialRefundBoundary(t *testing.T) {
	// 已支付 10000，已退 8000，本次退 2000 → 刚好退完，合法
	if err := ValidateRefund(model.PaymentStatusPaid, 10000, 8000, 2000); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

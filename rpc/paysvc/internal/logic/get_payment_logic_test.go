package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestGetPayment_ByPaymentNo(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 25.00, "wechat", 2, "tx_1", 5.00, nil, time.Now(), time.Now()))

	l := NewGetPaymentLogic(context.Background(), svcCtx)
	resp, err := l.GetPayment(&proto.GetPaymentRequest{PaymentNo: "PAY123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PaymentNo != "PAY123" {
		t.Errorf("payment_no = %s, want PAY123", resp.PaymentNo)
	}
	if resp.AmountCents != 2500 {
		t.Errorf("amount = %d, want 2500", resp.AmountCents)
	}
	if resp.Status != 2 {
		t.Errorf("status = %d, want 2", resp.Status)
	}
	if resp.RefundAmountCents != 500 {
		t.Errorf("refund = %d, want 500", resp.RefundAmountCents)
	}
}

func TestGetPayment_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY404", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	l := NewGetPaymentLogic(context.Background(), svcCtx)
	_, err := l.GetPayment(&proto.GetPaymentRequest{PaymentNo: "PAY404"})
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRefundPayment_Success(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 1. 查询支付单：已支付 status=2，金额 100 元，未退款
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 100.00, "wechat", 2, "tx_1", 0.00, nil, time.Now(), time.Now()))

	// 2. 更新支付单（部分退款：退款 3000 分 = 30 元）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	l := NewRefundPaymentLogic(context.Background(), svcCtx)
	resp, err := l.RefundPayment(&proto.RefundPaymentRequest{
		PaymentNo:         "PAY123",
		RefundAmountCents: 3000,
		Reason:            "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.RefundedAmountCents != 3000 {
		t.Errorf("refunded = %d, want 3000", resp.RefundedAmountCents)
	}
}

func TestRefundPayment_Exceed(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 已支付 100 元，退款 300 元 → 超额
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 100.00, "wechat", 2, "tx_1", 0.00, nil, time.Now(), time.Now()))

	l := NewRefundPaymentLogic(context.Background(), svcCtx)
	_, err := l.RefundPayment(&proto.RefundPaymentRequest{
		PaymentNo:         "PAY123",
		RefundAmountCents: 30000,
	})
	if !errors.Is(err, rule.ErrRefundExceed) {
		t.Errorf("expected ErrRefundExceed, got %v", err)
	}
}

func TestRefundPayment_NotPaid(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 待支付状态 status=1，不能退款
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 100.00, "wechat", 1, "", 0.00, nil, time.Now(), time.Now()))

	l := NewRefundPaymentLogic(context.Background(), svcCtx)
	_, err := l.RefundPayment(&proto.RefundPaymentRequest{
		PaymentNo:         "PAY123",
		RefundAmountCents: 1000,
	})
	if !errors.Is(err, rule.ErrRefundNotAllowed) {
		t.Errorf("expected ErrRefundNotAllowed, got %v", err)
	}
}

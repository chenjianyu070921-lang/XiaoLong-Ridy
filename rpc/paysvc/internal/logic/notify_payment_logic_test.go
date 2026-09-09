package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

// paymentColumns 与 model.Payment 的 GORM 列对应。
var paymentColumns = []string{
	"id", "payment_no", "order_id", "user_id", "amount", "channel",
	"status", "transaction_id", "refund_amount", "paid_at", "created_at", "updated_at",
}

func TestNotifyPayment_VerifyFail(t *testing.T) {
	db, _ := newMockDB(t)
	verifier := &channel.MockVerifier{Err: errors.New("bad sign")}
	svcCtx := newTestSvcCtx(db, nil, verifier)

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	resp, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:   "PAY123",
		TradeStatus: alipayTradeSuccess,
		NotifyRaw:   "x=1",
	})
	if err == nil {
		t.Fatal("expected verify error")
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
}

func TestNotifyPayment_AlreadyPaid(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 25.00, "wechat", 2, "tx_1", 0.00, nil, time.Now(), time.Now()))

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	resp, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:   "PAY123",
		TradeStatus: alipayTradeSuccess,
		NotifyRaw:   "x=1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestNotifyPayment_SuccessFullChain(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, &mockOrderClient{driverId: 3001}, nil)

	// 1. 查询支付单（待支付 status=1）
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 25.00, "wechat", 1, "", 0.00, nil, time.Now(), time.Now()))

	// 2. 更新支付单为支付成功（Save 包事务）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// 3. 触发结算：INSERT settlement（Create 包事务）
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `settlement`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	resp, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:     "PAY123",
		TradeStatus:   alipayTradeSuccess,
		TransactionId: "tx_123",
		PaidAt:        1753065600,
		NotifyRaw:     "x=1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestNotifyPayment_IgnoreNonSuccessStatus(t *testing.T) {
	db, _ := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	resp, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:   "PAY123",
		TradeStatus: "TRADE_CLOSED", // 非成功状态
		NotifyRaw:   "x=1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success (ignore)")
	}
}

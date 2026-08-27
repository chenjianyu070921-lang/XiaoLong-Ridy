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
	"status", "transaction_id", "refund_amount", "event_sent", "paid_at", "created_at", "updated_at",
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

	// 幂等路径：进入事务 → SELECT → 直接 commit（不更新）。
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 2500, "wechat", 2, "tx_1", 0, 0, nil, time.Now(), time.Now()))
	mock.ExpectCommit()

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
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 1. 事务开始 + 读取支付单（待支付 status=1）
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 2500, "wechat", 1, "", 0, 0, nil, time.Now(), time.Now()))

	// 2. 条件更新 → GORM 用条件 UPDATE 而不是事务 SAVE
	mock.ExpectExec("UPDATE `payment` SET").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// 3. 事件发送成功后，更新 event_sent=true（对账标记）。
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment` SET").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	resp, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:        "PAY123",
		TradeStatus:      alipayTradeSuccess,
		TransactionId:    "tx_123",
		PaidAt:           1753065600,
		NotifyRaw:        "x=1",
		TotalAmountCents: 2500, // 与 amount(2500) 一致
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

// TestNotifyPayment_AmountMismatch 验签通过但回调金额与支付单不一致：必须拒绝。
func TestNotifyPayment_AmountMismatch(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs("PAY123", 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY123", 1001, 2001, 2500, "wechat", 1, "", 0, 0, nil, time.Now(), time.Now()))
	// 金额比对失败 → 事务回滚，不发 UPDATE。
	mock.ExpectRollback()

	l := NewNotifyPaymentLogic(context.Background(), svcCtx)
	_, err := l.NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:        "PAY123",
		TradeStatus:      alipayTradeSuccess,
		NotifyRaw:        "x=1",
		TotalAmountCents: 9999, // 与 2500(=2500 cents) 不一致
	})
	if err == nil {
		t.Fatal("expected amount mismatch error")
	}
	if !errors.Is(err, ErrPaymentAmountMismatch) {
		t.Errorf("expected ErrPaymentAmountMismatch, got %v", err)
	}
}

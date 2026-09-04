package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestDetectDiff_ChannelMissing 平台已支付但渠道查不到 → 记为"渠道缺失"差异。
func TestDetectDiff_ChannelMissing(t *testing.T) {
	p := &model.Payment{
		PaymentNo:   "PAY_TEST_001",
		OrderId:     1001,
		AmountCents: 2950,
		Status:      model.PaymentStatusPaid,
	}
	diff := detectDiff(p, nil)
	if diff == nil {
		t.Fatal("want a diff when channel missing, got nil")
	}
	if diff.DiffType != model.ReconcileDiffChannelOnly {
		t.Fatalf("want diff type %d, got %d", model.ReconcileDiffChannelOnly, diff.DiffType)
	}
	if diff.PlatformAmount != 2950 {
		t.Fatalf("want platform amount 2950, got %d", diff.PlatformAmount)
	}
}

// TestDetectDiff_AmountMismatch 渠道金额与平台不一致 → 记为"金额不一致"差异。
func TestDetectDiff_AmountMismatch(t *testing.T) {
	p := &model.Payment{
		PaymentNo:   "PAY_TEST_002",
		OrderId:     1002,
		AmountCents: 2950,
		Status:      model.PaymentStatusPaid,
	}
	ch := &ChannelTransaction{
		TransactionId: "ALIPAY_TX_002",
		AmountCents:   2951, // 差 1 分
		Status:        "TRADE_SUCCESS",
	}
	diff := detectDiff(p, ch)
	if diff == nil {
		t.Fatal("want a diff on amount mismatch, got nil")
	}
	if diff.DiffType != model.ReconcileDiffAmount {
		t.Fatalf("want diff type %d, got %d", model.ReconcileDiffAmount, diff.DiffType)
	}
}

// TestDetectDiff_NoDiff 渠道金额与平台一致 → 无差异。
func TestDetectDiff_NoDiff(t *testing.T) {
	p := &model.Payment{
		PaymentNo:   "PAY_TEST_003",
		OrderId:     1003,
		AmountCents: 2950,
		Status:      model.PaymentStatusPaid,
	}
	ch := &ChannelTransaction{
		TransactionId: "ALIPAY_TX_003",
		AmountCents:   2950,
		Status:        "TRADE_SUCCESS",
	}
	if diff := detectDiff(p, ch); diff != nil {
		t.Fatalf("want nil diff, got %v", diff)
	}
}

// TestChannelReconcileOnce 对账 job 一次执行：扫描→差异写入→执行日志完成。
func TestChannelReconcileOnce(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewPaymentReconcileRepo(db)
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 1. 创建 run log（INSERT）
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment_channel_reconcile_log`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 2. 扫描已支付单（SELECT）：返回 1 笔已支付单
	paidAt := time.Now().Add(-time.Minute)
	mock.ExpectQuery("SELECT .* FROM `payment`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_no", "order_id", "amount", "status", "paid_at", "channel", "transaction_id", "refund_amount", "event_sent", "created_at", "updated_at"}).
			AddRow(1, "PAY_TEST_001", 1001, 2950, model.PaymentStatusPaid, paidAt, "alipay", "TX001", 0, true, time.Now(), time.Now()))

	// 3. QueryChannelTransactions（结构骨架返回空 map，不涉及 SQL）

	// 4. 写差异（INSERT）——渠道缺失
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment_reconcile_diff`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 5. 结束执行日志（UPDATE）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment_channel_reconcile_log`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	channelReconcileOnce(context.Background(), svcCtx, repo, time.Hour)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

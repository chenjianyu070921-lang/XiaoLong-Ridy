package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestPaymentFlowE2E 端到端串联整条支付正向链路：
//
//	创建支付单(CreatePayment) → 支付回调成功(NotifyPayment) → 触发司机结算(SettleOrder)
//	  → 支付查询(GetPayment) → 退款(RefundPayment)
//
// 全程使用 sqlmock 内存断言 SQL + MockChannel + NoopProducer，不依赖 MySQL/Kafka。
func TestPaymentFlowE2E(t *testing.T) {
	db, mock := newMockDB(t)
	// driver_id=3001：回调成功后 settleAfterPaid 会走 GetDriverId 拿到司机并触发结算
	svcCtx := newTestSvcCtx(db, &mockOrderClient{driverId: 3001}, nil)
	ctx := context.Background()

	// ============ 1. CreatePayment ============
	// repo.Create → INSERT payment（GORM 自动事务）
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// 回填 transaction_id → repo.Update → Save → UPDATE payment（事务）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	createResp, err := NewCreatePaymentLogic(ctx, svcCtx).CreatePayment(&proto.CreatePaymentRequest{
		OrderId:     1001,
		UserId:      2001,
		AmountCents: 2500,
		Channel:     proto.PayChannel_PAY_CHANNEL_WECHAT,
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if createResp.PaymentNo == "" {
		t.Fatal("CreatePayment: payment_no should not be empty")
	}
	if createResp.TransactionId == "" {
		t.Fatal("CreatePayment: transaction_id should not be empty")
	}
	if createResp.Status != 1 {
		t.Fatalf("CreatePayment: status = %d, want 1", createResp.Status)
	}

	// ============ 2. NotifyPayment（支付成功回调） ============
	// FindByPaymentNo → SELECT（payment_no 运行时生成，用 AnyArg 匹配）
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY_ANY", 1001, 2001, 25.00, "wechat", 1, "", 0.00, nil, time.Now(), time.Now()))
	// 更新为「支付成功」→ Save → UPDATE（事务）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// settleAfterPaid → SettleOrder → INSERT settlement（事务）
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `settlement`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	notifyResp, err := NewNotifyPaymentLogic(ctx, svcCtx).NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:     createResp.PaymentNo,
		TradeStatus:   alipayTradeSuccess,
		TransactionId: "tx_e2e",
		PaidAt:        1753065600,
		NotifyRaw:     "x=1",
	})
	if err != nil {
		t.Fatalf("NotifyPayment: %v", err)
	}
	if !notifyResp.Success {
		t.Fatal("NotifyPayment: expected success")
	}

	// ============ 3. GetPayment（查询应显示已支付） ============
	paidAt := time.Unix(1753065600, 0)
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY_ANY", 1001, 2001, 25.00, "wechat", 2, "tx_e2e", 0.00, &paidAt, time.Now(), time.Now()))

	getResp, err := NewGetPaymentLogic(ctx, svcCtx).GetPayment(&proto.GetPaymentRequest{
		PaymentNo: createResp.PaymentNo,
	})
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if getResp.Status != 2 {
		t.Fatalf("GetPayment: status = %d, want 2(paid)", getResp.Status)
	}
	if getResp.AmountCents != 2500 {
		t.Fatalf("GetPayment: amount = %d, want 2500", getResp.AmountCents)
	}

	// ============ 4. RefundPayment（部分退款 1000 分） ============
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY_ANY", 1001, 2001, 25.00, "wechat", 2, "tx_e2e", 0.00, &paidAt, time.Now(), time.Now()))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	refundResp, err := NewRefundPaymentLogic(ctx, svcCtx).RefundPayment(&proto.RefundPaymentRequest{
		PaymentNo:         createResp.PaymentNo,
		RefundAmountCents: 1000,
		Reason:            "e2e test",
	})
	if err != nil {
		t.Fatalf("RefundPayment: %v", err)
	}
	if !refundResp.Success {
		t.Fatal("RefundPayment: expected success")
	}
	if refundResp.RefundedAmountCents != 1000 {
		t.Fatalf("RefundPayment: refunded = %d, want 1000", refundResp.RefundedAmountCents)
	}

	// 所有 mock 期望都应被消费
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

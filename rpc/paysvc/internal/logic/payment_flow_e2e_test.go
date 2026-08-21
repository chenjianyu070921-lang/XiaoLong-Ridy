package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestPaymentFlowE2E 端到端串联支付正向链路：
//
//	创建支付单(CreatePayment) → 支付回调成功(NotifyPayment) → 支付查询(GetPayment) → 退款(RefundPayment)
//
// 注意：原来的 settleAfterPaid 链路已经改成异步消息驱动（M5-6），
// 结算由独立的 consumer 模块处理，不在本测试内串联。
// 事件投递由 NoopProducer 兜底（消息丢弃不影响业务正确性）。
//
// 全程使用 sqlmock 内存断言 SQL + MockChannel + NoopProducer，不依赖 MySQL/Kafka。
func TestPaymentFlowE2E(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, &mockOrderClient{driverId: 3001}, nil)
	ctx := context.Background()

	// ============ 1. CreatePayment（事务 + 渠道下单） ============
	// 渠道下单是事务外的 RPC；这里用 MockChannel 不触发外部 IO，仅校验：
	//   - 事务 Begin
	//   - INSERT payment（GORM 自动 Begin/Commit）
	//   - 事务 Commit
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment`").WillReturnResult(sqlmock.NewResult(1, 1))
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

	// ============ 2. NotifyPayment（支付成功回调 + 事务条件更新） ============
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY_ANY", 1001, 2001, 25.00, "wechat", 1, "", 0.00, nil, time.Now(), time.Now()))
	mock.ExpectExec("UPDATE `payment` SET").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	notifyResp, err := NewNotifyPaymentLogic(ctx, svcCtx).NotifyPayment(&proto.NotifyPaymentRequest{
		PaymentNo:        createResp.PaymentNo,
		TradeStatus:      alipayTradeSuccess,
		TransactionId:    "tx_e2e",
		PaidAt:           1753065600,
		NotifyRaw:        "x=1",
		TotalAmountCents: 2500, // 与 amount(25.00) 一致
	})
	if err != nil {
		t.Fatalf("NotifyPayment: %v", err)
	}
	if !notifyResp.Success {
		t.Fatal("NotifyPayment: expected success")
	}

	// ============ 3. GetPayment（事务外查询应显示已支付） ============
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

	// ============ 4. RefundPayment（部分退款 1000 分，事务条件更新） ============
	mock.ExpectQuery("SELECT \\* FROM `payment` WHERE payment_no = \\?").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(paymentColumns).
			AddRow(1, "PAY_ANY", 1001, 2001, 25.00, "wechat", 2, "tx_e2e", 0.00, &paidAt, time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment` SET").
		WillReturnResult(sqlmock.NewResult(0, 1))
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

	// 确认 NoopProducer 是真的不发消息（验证 M5-6 异步路径生效）。
	if _, ok := svcCtx.Producer.(*mq.NoopProducer); !ok {
		t.Errorf("expected NoopProducer, got %T", svcCtx.Producer)
	}

	// 所有 mock 期望都应被消费
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

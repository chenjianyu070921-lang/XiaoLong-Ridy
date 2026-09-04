package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestAutoSettleOnce 自动结算 job 一次执行：扫描未结算已支付单 → 调 SettleOrder 生成结算单 → 打标 auto_settled。
func TestAutoSettleOnce(t *testing.T) {
	db, mock := newMockDB(t)
	payRepo := repository.NewPaymentRepo(db)

	oc := &mockOrderClient{driverId: 888}
	svcCtx := newTestSvcCtx(db, oc, nil)

	// 1. 扫描已支付未结算单（JOIN）
	paidAt := time.Now().Add(-time.Minute)
	mock.ExpectQuery(`SELECT \* FROM payment AS p LEFT JOIN settlement AS s ON s\.order_id = p\.order_id WHERE p\.status = \? AND s\.id IS NULL ORDER BY p\.id ASC LIMIT \?`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_no", "order_id", "user_id", "amount", "status", "paid_at", "channel", "transaction_id", "refund_amount", "event_sent", "created_at", "updated_at"}).
			AddRow(1, "PAY_TEST_010", 2001, 3001, 5000, model.PaymentStatusPaid, paidAt, "alipay", "TX010", 0, true, time.Now(), time.Now()))

	// 2. SettleOrderLogic 内事务：INSERT settlement
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `settlement`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 3. 打标 auto_settled=1 + run_id（UPDATE settlement）
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `settlement`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	autoSettleOnce(context.Background(), svcCtx, payRepo, 20)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

// TestAutoSettleOnce_OrderClientError 订单客户端报错 → 跳过该笔，不 panic。
func TestAutoSettleOnce_OrderClientError(t *testing.T) {
	db, mock := newMockDB(t)
	payRepo := repository.NewPaymentRepo(db)

	oc := &mockOrderClient{err: errors.New("order not found")}
	svcCtx := newTestSvcCtx(db, oc, nil)

	paidAt := time.Now().Add(-time.Minute)
	mock.ExpectQuery(`SELECT \* FROM payment AS p LEFT JOIN settlement AS s ON s\.order_id = p\.order_id WHERE p\.status = \? AND s\.id IS NULL ORDER BY p\.id ASC LIMIT \?`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_no", "order_id", "user_id", "amount", "status", "paid_at", "channel", "transaction_id", "refund_amount", "event_sent", "created_at", "updated_at"}).
			AddRow(1, "PAY_TEST_011", 2002, 3002, 5000, model.PaymentStatusPaid, paidAt, "alipay", "TX011", 0, true, time.Now(), time.Now()))

	// GetDriverId 报错 → 不产生任何 DB 写操作。
	autoSettleOnce(context.Background(), svcCtx, payRepo, 20)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

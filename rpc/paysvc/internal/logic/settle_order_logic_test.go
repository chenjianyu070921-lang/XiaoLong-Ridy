package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSettleOrder_Success(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newTestSvcCtx(db, nil, nil)

	// 创建结算单（Create 包事务）
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `settlement`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	l := NewSettleOrderLogic(context.Background(), svcCtx)
	resp, err := l.SettleOrder(&proto.SettleOrderRequest{
		OrderId:          1001,
		DriverId:         3001,
		TotalAmountCents: 10000,
		CommissionRate:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10000 分，抽成 20% → 平台 2000，司机 8000
	if resp.PlatformCommissionCents != 2000 {
		t.Errorf("commission = %d, want 2000", resp.PlatformCommissionCents)
	}
	if resp.DriverIncomeCents != 8000 {
		t.Errorf("income = %d, want 8000", resp.DriverIncomeCents)
	}
	if resp.SettlementNo == "" {
		t.Error("settlement_no should not be empty")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

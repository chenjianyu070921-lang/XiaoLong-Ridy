package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newMockDB 创建基于 sqlmock 的 GORM DB。
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db, mock
}

// orderPriceColumns 与 model.OrderPrice 的 GORM 列对应。
var orderPriceColumns = []string{
	"id", "order_id", "price_rule_id", "estimated_price", "actual_price",
	"base_fee", "distance_fee", "time_fee", "night_fee", "dynamic_fee",
	"discount_amount", "platform_subsidy", "payable_amount", "status",
	"created_at", "updated_at",
}

func newPriceSvcCtx(db *gorm.DB) *svc.ServiceContext {
	return &svc.ServiceContext{DB: db}
}

func TestSaveActualOrderPrice_UpdateExisting(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	// 整个调用都在一个事务里 → Begin → SELECT → UPDATE → Commit。
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `order_price` WHERE order_id = \\?").
		WithArgs(1001, 1).
		WillReturnRows(sqlmock.NewRows(orderPriceColumns).
			AddRow(1, 1001, 1, 25.00, 0.00, 8.00, 10.00, 4.00, 0.00, 0.00, 0.00, 0.00, 0.00, 1, time.Now(), time.Now()))
	mock.ExpectExec("UPDATE `order_price` SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	l := NewSaveActualOrderPriceLogic(ctx, svcCtx)
	resp, err := l.SaveActualOrderPrice(&proto.SaveActualOrderPriceRequest{
		OrderId:          1001,
		PriceRuleId:      1,
		ActualPriceCents: 3000,
		Detail: &proto.PriceDetail{
			BaseFeeCents:     800,
			DistanceFeeCents: 1500,
			TimeFeeCents:     500,
			TotalCents:       3000,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.OrderPriceId != 1 {
		t.Errorf("order_price_id = %d, want 1", resp.OrderPriceId)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSaveActualOrderPrice_CreateNew(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	// 整个调用都在事务里 → Begin → SELECT(not found) → INSERT → Commit。
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `order_price` WHERE order_id = \\?").
		WithArgs(1002, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec("INSERT INTO `order_price`").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	l := NewSaveActualOrderPriceLogic(ctx, svcCtx)
	resp, err := l.SaveActualOrderPrice(&proto.SaveActualOrderPriceRequest{
		OrderId:          1002,
		PriceRuleId:      2,
		ActualPriceCents: 2000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.OrderPriceId != 2 {
		t.Errorf("order_price_id = %d, want 2", resp.OrderPriceId)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSaveActualOrderPrice_InvalidOrderId(t *testing.T) {
	db, _ := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	l := NewSaveActualOrderPriceLogic(ctx, svcCtx)
	_, err := l.SaveActualOrderPrice(&proto.SaveActualOrderPriceRequest{
		OrderId: 0,
	})
	if err != ErrOrderIdInvalid {
		t.Errorf("expected ErrOrderIdInvalid, got %v", err)
	}
}

func TestSaveActualOrderPrice_InvalidPrice(t *testing.T) {
	db, _ := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	l := NewSaveActualOrderPriceLogic(ctx, svcCtx)
	_, err := l.SaveActualOrderPrice(&proto.SaveActualOrderPriceRequest{
		OrderId:          1003,
		ActualPriceCents: -1,
	})
	if err != ErrActualPriceInvalid {
		t.Errorf("expected ErrActualPriceInvalid, got %v", err)
	}
}

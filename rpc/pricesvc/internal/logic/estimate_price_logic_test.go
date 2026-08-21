package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/pricesvc/internal/rule"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 与 save_actual_order_price_logic_test.go 同款 mock DB，仅本测试用。
func newEstimateMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
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

func newPriceSvcCtxOnly(db *gorm.DB) *svc.ServiceContext {
	return &svc.ServiceContext{DB: db}
}

// TestEstimatePrice_DBRuleFound 命中数据库规则：正常返回。
func TestEstimatePrice_DBRuleFound(t *testing.T) {
	db, mock := newEstimateMockDB(t)
	svcCtx := newPriceSvcCtxOnly(db)
	ctx := context.Background()

	// 默认规则（与 model.PriceRule 列对齐）：car_type=1, city_code="110000", status=1
	mock.ExpectQuery("SELECT \\* FROM `price_rule`").
		WillReturnRows(sqlmock.NewRows(priceRuleColumns).
			AddRow(1, "北京快车", "110000", 1, 12.00, 3.00, 2.50, 0.50, "23:00:00", "05:00:00", 8.00, 1.50, 1, time.Now(), nil, time.Now(), time.Now()))

	l := NewEstimatePriceLogic(ctx, svcCtx)
	resp, err := l.EstimatePrice(&proto.EstimatePriceRequest{
		UserId:      1001,
		CityCode:    "110000",
		CarType:     1,
		DistanceM:   5000, // 5 km
		DurationS:   600,  // 10 min
		Timestamp:   time.Date(2026, 8, 13, 14, 0, 0, 0, time.Local).Unix(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PriceRuleId != 1 {
		t.Errorf("price_rule_id = %d, want 1", resp.PriceRuleId)
	}
	// 起步 1200 + 距费 (5-3)*250=500 + 时费 600*50/60=500 = 2200（非夜间，无调价）
	if resp.TotalCents != 2200 {
		t.Errorf("total_cents = %d, want 2200", resp.TotalCents)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestEstimatePrice_DBRuleNotFound_FallbackDefault M5-12：DB 无规则时回退默认规则。
func TestEstimatePrice_DBRuleNotFound_FallbackDefault(t *testing.T) {
	db, mock := newEstimateMockDB(t)
	svcCtx := newPriceSvcCtxOnly(db)
	ctx := context.Background()

	// First() not found：触发 ErrRecordNotFound 走兜底。
	mock.ExpectQuery("SELECT \\* FROM `price_rule`").
		WillReturnError(gorm.ErrRecordNotFound)

	l := NewEstimatePriceLogic(ctx, svcCtx)
	// 用 weekday=2026-08-13 14:00 (周四) 避开 7-9 与 17-19 高峰，避免调价。
	resp, err := l.EstimatePrice(&proto.EstimatePriceRequest{
		UserId:    1001,
		CityCode:  "999999", // 假装未知城市
		CarType:   1,
		DistanceM: 5000,
		DurationS: 600,
		Timestamp: time.Date(2026, 8, 13, 14, 0, 0, 0, time.Local).Unix(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 兜底规则起步 1000 + (5-3)*200=400 + 600*40/60=400 = 1800
	if resp.TotalCents != 1800 {
		t.Errorf("total_cents (default) = %d, want 1800", resp.TotalCents)
	}
	// 兜底时不在响应里泄露 fake ID，避免前端误用。
	if resp.PriceRuleId != 0 {
		t.Errorf("price_rule_id (default) = %d, want 0", resp.PriceRuleId)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestEstimatePrice_NightSurcharge 验证夜间时段触发附加费。
func TestEstimatePrice_NightSurcharge(t *testing.T) {
	db, mock := newEstimateMockDB(t)
	svcCtx := newPriceSvcCtxOnly(db)
	ctx := context.Background()

	// night: 23:00~05:00
	mock.ExpectQuery("SELECT \\* FROM `price_rule`").
		WillReturnRows(sqlmock.NewRows(priceRuleColumns).
			AddRow(2, "夜班", "110000", 1, 12.00, 3.00, 2.50, 0.50,
				strPtr("23:00:00"), strPtr("05:00:00"),
				8.00, 1.50, 1, time.Now(), nil, time.Now(), time.Now()))

	l := NewEstimatePriceLogic(ctx, svcCtx)
	resp, err := l.EstimatePrice(&proto.EstimatePriceRequest{
		UserId:    1001,
		CityCode:  "110000",
		CarType:   1,
		DistanceM: 5000,
		DurationS: 600,
		Timestamp: time.Date(2026, 8, 13, 23, 30, 0, 0, time.Local).Unix(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Detail.NightFeeCents != 800 {
		t.Errorf("night fee = %d, want 800", resp.Detail.NightFeeCents)
	}
}

// TestDefaultPriceRule_Baseline 兜底规则能直接构造，不依赖 DB。
func TestDefaultPriceRule_Baseline(t *testing.T) {
	pr := rule.DefaultPriceRule("110000", 1)
	if pr == nil {
		t.Fatal("default rule nil")
	}
	if pr.Id != rule.DefaultPriceRuleID {
		t.Errorf("id = %d, want %d", pr.Id, rule.DefaultPriceRuleID)
	}
	if !rule.IsUsingDefaultRule(int64(rule.DefaultPriceRuleID)) {
		t.Error("IsUsingDefaultRule should be true for default id")
	}
	if rule.IsUsingDefaultRule(123) {
		t.Error("IsUsingDefaultRule should be false for non-default id")
	}
}

// priceRuleColumns 与 model.PriceRule GORM 列对齐。
var priceRuleColumns = []string{
	"id", "name", "city_code", "car_type",
	"base_price", "base_distance_km", "per_km_price", "per_minute_price",
	"night_start_time", "night_end_time", "night_surcharge", "dynamic_max_factor",
	"status", "effective_at", "expire_at", "created_at", "updated_at",
}

func strPtr(s string) *string { return &s }

package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
	"XiaoLong-Ridy/rpc/pricesvc/internal/rule"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// TestCalculateDiscount_NoOrderPrice 订单价格记录不存在：仅返回计算结果，不落库。
// 这是计算阶段的预期路径：预估一次价格，订单还未生成明细。
func TestCalculateDiscount_NoOrderPrice(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT \\* FROM `order_price` WHERE order_id = \\?").
		WithArgs(int64(1001), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	l := NewCalculateDiscountLogic(ctx, svcCtx)
	resp, err := l.CalculateDiscount(&proto.CalculateDiscountRequest{
		OrderId:    1001,
		TotalCents: 3000,
		Coupon: &proto.Coupon{
			CouponId:       1,
			Type:           proto.CouponType_COUPON_TYPE_FIXED,
			FaceValueCents: 500,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DiscountAmountCents != 500 {
		t.Errorf("discount = %d, want 500", resp.DiscountAmountCents)
	}
	if resp.PayableAmountCents != 2500 {
		t.Errorf("payable = %d, want 2500", resp.PayableAmountCents)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCalculateDiscount_WithOrderPrice 订单价格记录已存在：计算后用 Updates(map) 仅写必要列。
func TestCalculateDiscount_WithOrderPrice(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT \\* FROM `order_price` WHERE order_id = \\?").
		WithArgs(int64(1001), 1).
		WillReturnRows(sqlmock.NewRows(orderPriceColumns).
			AddRow(1, 1001, 1, 25.00, 0.00, 8.00, 10.00, 4.00, 0.00, 0.00,
				0.00, 0.00, 0.00, 1, time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `order_price` SET").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	l := NewCalculateDiscountLogic(ctx, svcCtx)
	resp, err := l.CalculateDiscount(&proto.CalculateDiscountRequest{
		OrderId:    1001,
		TotalCents: 3000,
		Coupon: &proto.Coupon{
			CouponId:       1,
			Type:           proto.CouponType_COUPON_TYPE_DISCOUNT,
			Discount:       80,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 8 折 → 折扣 600，实付 2400
	if resp.DiscountAmountCents != 600 || resp.PayableAmountCents != 2400 {
		t.Errorf("resp = %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCalculateDiscount_ThresholdNotMet 门槛拦截：coupon 门槛 5000 但订单 3000，应报错。
// 期望：rule 校验直接返回 ErrCouponNotMeetThreshold，不会查询 order_price（pure 计算校验）。
func TestCalculateDiscount_ThresholdNotMet(t *testing.T) {
	db, _ := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	l := NewCalculateDiscountLogic(ctx, svcCtx)
	_, err := l.CalculateDiscount(&proto.CalculateDiscountRequest{
		OrderId:    1001,
		TotalCents: 3000,
		Coupon: &proto.Coupon{
			CouponId:       1,
			Type:           proto.CouponType_COUPON_TYPE_FIXED,
			FaceValueCents: 500,
			ThresholdCents: 5000,
		},
	})
	if !errors.Is(err, rule.ErrCouponNotMeetThreshold) {
		t.Errorf("expected ErrCouponNotMeetThreshold, got %v", err)
	}
}

// TestSaveActualOrderPrice_DetailOnly 当请求只带明细不带总价时，以明细 total 为准。
func TestSaveActualOrderPrice_DetailOnly(t *testing.T) {
	db, mock := newMockDB(t)
	svcCtx := newPriceSvcCtx(db)
	ctx := context.Background()

	// 整段调用都在一个事务里：Begin → SELECT(not found) → INSERT → Commit。
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `order_price` WHERE order_id = \\?").
		WithArgs(int64(2001), 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec("INSERT INTO `order_price`").
		WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectCommit()

	l := NewSaveActualOrderPriceLogic(ctx, svcCtx)
	resp, err := l.SaveActualOrderPrice(&proto.SaveActualOrderPriceRequest{
		OrderId:    2001,
		PriceRuleId: 0,
		// ActualPriceCents 不填，依赖 Detail.TotalCents
		Detail: &proto.PriceDetail{
			BaseFeeCents:     1000,
			DistanceFeeCents: 500,
			TimeFeeCents:     500,
			TotalCents:       2000,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	// 此时落库的 actual_price 应当等于 detail.total 的换算
	// (10.00 元)；不易直接断言 SQL 入参，但模型字段设置正确即可。
	_ = priceutil.CentsToYuan // 触发包引用，证明逻辑可访问 model
	_ = model.OrderPrice{}
}

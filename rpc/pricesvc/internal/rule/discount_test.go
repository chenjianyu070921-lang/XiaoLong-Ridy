package rule

import "testing"

func TestCalculateDiscount_Fixed(t *testing.T) {
	// 订单 3000 分，固定券面额 500 分
	res, err := CalculateDiscount(3000, CouponInput{
		Type:           CouponTypeFixed,
		FaceValueCents: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscountAmountCents != 500 {
		t.Errorf("discount = %d, want 500", res.DiscountAmountCents)
	}
	if res.PayableAmountCents != 2500 {
		t.Errorf("payable = %d, want 2500", res.PayableAmountCents)
	}
}

func TestCalculateDiscount_FixedExceedsTotal(t *testing.T) {
	// 面额 500 大于订单 300，折扣应封顶为 300
	res, err := CalculateDiscount(300, CouponInput{
		Type:           CouponTypeFixed,
		FaceValueCents: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscountAmountCents != 300 {
		t.Errorf("discount = %d, want 300", res.DiscountAmountCents)
	}
	if res.PayableAmountCents != 0 {
		t.Errorf("payable = %d, want 0", res.PayableAmountCents)
	}
}

func TestCalculateDiscount_DiscountRate(t *testing.T) {
	// 8 折（discount=80），订单 3000 → 优惠 20% = 600
	res, err := CalculateDiscount(3000, CouponInput{
		Type:     CouponTypeDiscount,
		Discount: 80,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscountAmountCents != 600 {
		t.Errorf("discount = %d, want 600", res.DiscountAmountCents)
	}
	if res.PayableAmountCents != 2400 {
		t.Errorf("payable = %d, want 2400", res.PayableAmountCents)
	}
}

func TestCalculateDiscount_MaxDiscount(t *testing.T) {
	// 8 折，最大优惠 500，订单 3000 → 理论优惠 600，但封顶 500
	res, err := CalculateDiscount(3000, CouponInput{
		Type:            CouponTypeDiscount,
		Discount:        80,
		MaxDiscountCents: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscountAmountCents != 500 {
		t.Errorf("discount = %d, want 500", res.DiscountAmountCents)
	}
	if res.PayableAmountCents != 2500 {
		t.Errorf("payable = %d, want 2500", res.PayableAmountCents)
	}
}

func TestCalculateDiscount_Threshold(t *testing.T) {
	// 门槛 5000，订单 3000 不满足
	if _, err := CalculateDiscount(3000, CouponInput{
		Type:           CouponTypeFixed,
		FaceValueCents: 500,
		ThresholdCents: 5000,
	}); err == nil {
		t.Error("expected threshold error")
	}
}

func TestCalculateDiscount_ThresholdMet(t *testing.T) {
	// 门槛 5000，订单 6000 满足
	res, err := CalculateDiscount(6000, CouponInput{
		Type:           CouponTypeFixed,
		FaceValueCents: 500,
		ThresholdCents: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.PayableAmountCents != 5500 {
		t.Errorf("payable = %d, want 5500", res.PayableAmountCents)
	}
}

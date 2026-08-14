package rule

import "testing"

// 构造一个典型规则（北京快车，金额单位：分）
func testRule() PriceRuleInput {
	return PriceRuleInput{
		BasePriceCents:     1200, // 起步价 12 元
		BaseDistanceM:      3000, // 起步包含 3 公里
		PerKmPriceCents:    250,  // 每公里 2.5 元
		PerMinutePriceCents: 50,  // 每分钟 0.5 元
		NightSurchargeCents: 800, // 夜间附加费 8 元
		DynamicMaxFactor:   1.5,  // 动态调价最大 1.5 倍
	}
}

func TestEstimate_BasicNoNight(t *testing.T) {
	r := testRule()
	// 5 公里、10 分钟、非夜间、无动态
	detail, err := Estimate(r, 5000, 600, false, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 里程费 = (5km - 3km) * 250 = 2km * 250 = 500
	if detail.DistanceFeeCents != 500 {
		t.Errorf("distance fee = %d, want 500", detail.DistanceFeeCents)
	}
	// 时长费 = 10min * 50 = 500
	if detail.TimeFeeCents != 500 {
		t.Errorf("time fee = %d, want 500", detail.TimeFeeCents)
	}
	// 起步价 1200
	if detail.BaseFeeCents != 1200 {
		t.Errorf("base fee = %d, want 1200", detail.BaseFeeCents)
	}
	// 夜间 0
	if detail.NightFeeCents != 0 {
		t.Errorf("night fee = %d, want 0", detail.NightFeeCents)
	}
	// 动态 0
	if detail.DynamicFeeCents != 0 {
		t.Errorf("dynamic fee = %d, want 0", detail.DynamicFeeCents)
	}
	// 总价 = 1200 + 500 + 500 = 2200
	if detail.TotalCents != 2200 {
		t.Errorf("total = %d, want 2200", detail.TotalCents)
	}
}

func TestEstimate_WithinBaseDistance(t *testing.T) {
	r := testRule()
	// 2 公里（小于起步 3 公里）、5 分钟
	detail, err := Estimate(r, 2000, 300, false, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 里程费 = 0（未超过起步里程）
	if detail.DistanceFeeCents != 0 {
		t.Errorf("distance fee = %d, want 0", detail.DistanceFeeCents)
	}
	// 总价 = 1200 + 0 + 250 = 1450
	if detail.TotalCents != 1450 {
		t.Errorf("total = %d, want 1450", detail.TotalCents)
	}
}

func TestEstimate_NightSurcharge(t *testing.T) {
	r := testRule()
	detail, err := Estimate(r, 5000, 600, true, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.NightFeeCents != 800 {
		t.Errorf("night fee = %d, want 800", detail.NightFeeCents)
	}
	// 总价 = 1200 + 500 + 500 + 800 = 3000
	if detail.TotalCents != 3000 {
		t.Errorf("total = %d, want 3000", detail.TotalCents)
	}
}

func TestEstimate_DynamicFactor(t *testing.T) {
	r := testRule()
	// 动态 1.2 倍：基础(起步+里程+时长) * 0.2 作为动态费
	detail, err := Estimate(r, 5000, 600, false, 1.2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 基础价 = 1200 + 500 + 500 = 2200
	// 动态费 = 2200 * (1.2 - 1) = 440
	if detail.DynamicFeeCents != 440 {
		t.Errorf("dynamic fee = %d, want 440", detail.DynamicFeeCents)
	}
	// 总价 = 2200 + 440 = 2640
	if detail.TotalCents != 2640 {
		t.Errorf("total = %d, want 2640", detail.TotalCents)
	}
}

func TestEstimate_DynamicFactorCapped(t *testing.T) {
	r := testRule()
	// 请求 3.0 倍，但规则上限 1.5 倍
	detail, err := Estimate(r, 5000, 600, false, 3.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 基础价 2200，动态费 = 2200 * (1.5 - 1) = 1100
	if detail.DynamicFeeCents != 1100 {
		t.Errorf("dynamic fee = %d, want 1100", detail.DynamicFeeCents)
	}
	// 总价 = 2200 + 1100 = 3300
	if detail.TotalCents != 3300 {
		t.Errorf("total = %d, want 3300", detail.TotalCents)
	}
}

func TestEstimate_InvalidInput(t *testing.T) {
	r := testRule()
	// 里程为负应报错
	if _, err := Estimate(r, -1, 600, false, 1.0); err == nil {
		t.Error("expected error for negative distance")
	}
	// 时长为负应报错
	if _, err := Estimate(r, 5000, -1, false, 1.0); err == nil {
		t.Error("expected error for negative duration")
	}
}

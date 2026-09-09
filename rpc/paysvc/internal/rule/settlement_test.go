package rule

import "testing"

func TestCalcSettlement_Basic(t *testing.T) {
	// 订单 10000 分（100 元），抽成 20%
	commission, income := CalcSettlement(10000, 20)
	if commission != 2000 {
		t.Errorf("commission = %d, want 2000", commission)
	}
	if income != 8000 {
		t.Errorf("income = %d, want 8000", income)
	}
}

func TestCalcSettlement_ZeroTotal(t *testing.T) {
	commission, income := CalcSettlement(0, 20)
	if commission != 0 || income != 0 {
		t.Errorf("expected 0,0 got %d,%d", commission, income)
	}
}

func TestCalcSettlement_RateBounds(t *testing.T) {
	// 抽成比例为负 → 按 0
	c1, i1 := CalcSettlement(10000, -5)
	if c1 != 0 || i1 != 10000 {
		t.Errorf("negative rate: got %d,%d want 0,10000", c1, i1)
	}
	// 抽成比例超 100 → 按 100
	c2, i2 := CalcSettlement(10000, 150)
	if c2 != 10000 || i2 != 0 {
		t.Errorf("over rate: got %d,%d want 10000,0", c2, i2)
	}
}

func TestCalcSettlement_Rounding(t *testing.T) {
	// 订单 333 分（3.33 元），抽成 15% → 49.95 分 → 四舍五入 50 分
	commission, income := CalcSettlement(333, 15)
	if commission != 50 {
		t.Errorf("commission = %d, want 50", commission)
	}
	if income != 283 {
		t.Errorf("income = %d, want 283", income)
	}
}

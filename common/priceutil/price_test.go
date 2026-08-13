package priceutil

import "testing"

func TestYuanToCents(t *testing.T) {
	cases := []struct {
		yuan float64
		want int64
	}{
		{1.23, 123},
		{0.00, 0},
		{8.00, 800},
		{12.345, 1235}, // 四舍五入到分
		{1.8, 180},
	}
	for _, c := range cases {
		got := YuanToCents(c.yuan)
		if got != c.want {
			t.Errorf("YuanToCents(%v) = %d, want %d", c.yuan, got, c.want)
		}
	}
}

func TestCentsToYuan(t *testing.T) {
	cases := []struct {
		cents int64
		want  float64
	}{
		{123, 1.23},
		{0, 0},
		{800, 8.00},
	}
	for _, c := range cases {
		got := CentsToYuan(c.cents)
		if got != c.want {
			t.Errorf("CentsToYuan(%d) = %v, want %v", c.cents, got, c.want)
		}
	}
}

func TestAdd(t *testing.T) {
	if got := Add(100, 200, 50); got != 350 {
		t.Errorf("Add(100,200,50) = %d, want 350", got)
	}
	if got := Add(); got != 0 {
		t.Errorf("Add() = %d, want 0", got)
	}
}

func TestMax(t *testing.T) {
	if got := Max(-5, 0); got != 0 {
		t.Errorf("Max(-5,0) = %d, want 0", got)
	}
	if got := Max(3, 10); got != 10 {
		t.Errorf("Max(3,10) = %d, want 10", got)
	}
}

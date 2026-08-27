package rule

import (
	"testing"
	"time"
)

func TestIsPeakTime(t *testing.T) {
	cases := []struct {
		h, m int
		want bool
	}{
		{6, 59, false}, // 早高峰前
		{7, 0, true},   // 早高峰开始
		{8, 59, true},  // 早高峰结束前
		{9, 0, false},  // 早高峰结束
		{12, 0, false}, // 平峰
		{16, 59, false},
		{17, 0, true},  // 晚高峰开始
		{18, 59, true}, // 晚高峰结束前
		{19, 0, false}, // 晚高峰结束
		{23, 0, false},
	}
	for _, c := range cases {
		tm := time.Date(2026, 8, 14, c.h, c.m, 0, 0, time.Local)
		if got := IsPeakTime(tm); got != c.want {
			t.Errorf("IsPeakTime(%02d:%02d) = %v, want %v", c.h, c.m, got, c.want)
		}
	}
}


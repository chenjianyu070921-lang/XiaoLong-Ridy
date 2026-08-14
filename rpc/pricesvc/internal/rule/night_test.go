package rule

import (
	"testing"
	"time"
)

func mustTime(h, m int) time.Time {
	return time.Date(2026, 8, 13, h, m, 0, 0, time.Local)
}

func TestIsNightTime_CrossMidnight(t *testing.T) {
	// 夜间 23:00 ~ 05:00
	cases := []struct {
		h, m int
		want bool
	}{
		{22, 59, false},
		{23, 0, true},
		{0, 0, true},
		{4, 59, true},
		{5, 0, false},
		{12, 0, false},
	}
	for _, c := range cases {
		got := IsNightTime(mustTime(c.h, c.m), "23:00:00", "05:00:00")
		if got != c.want {
			t.Errorf("IsNightTime(%02d:%02d) = %v, want %v", c.h, c.m, got, c.want)
		}
	}
}

func TestIsNightTime_SameDay(t *testing.T) {
	// 非跨天 20:00 ~ 22:00
	cases := []struct {
		h, m int
		want bool
	}{
		{19, 59, false},
		{20, 0, true},
		{21, 59, true},
		{22, 0, false},
	}
	for _, c := range cases {
		got := IsNightTime(mustTime(c.h, c.m), "20:00:00", "22:00:00")
		if got != c.want {
			t.Errorf("IsNightTime(%02d:%02d) = %v, want %v", c.h, c.m, got, c.want)
		}
	}
}

func TestIsNightTime_Empty(t *testing.T) {
	if IsNightTime(mustTime(23, 30), "", "") {
		t.Error("expected false for empty night time")
	}
	if IsNightTime(mustTime(23, 30), "", "05:00:00") {
		t.Error("expected false for empty start")
	}
}

package middleware

import (
	"net/http/httptest"
	"testing"
	"time"
)

// newFakeLimiter 构造注入假时钟的限流器，便于测试窗口重置。
func newFakeLimiter(limit int, windowSize time.Duration, now *time.Time) *loginLimiter {
	return &loginLimiter{
		buckets: make(map[string]*window),
		limit:   limit,
		window:  windowSize,
		now:     func() time.Time { return *now },
	}
}

func TestLimiterAllowWithinLimit(t *testing.T) {
	now := time.Now()
	l := newFakeLimiter(3, time.Minute, &now)
	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("allow(%d) = false, want true", i)
		}
	}
	// 第 4 次超限。
	if l.allow("1.2.3.4") {
		t.Fatal("allow(4) = true, want false")
	}
}

func TestLimiterDifferentKeysIsolated(t *testing.T) {
	now := time.Now()
	l := newFakeLimiter(1, time.Minute, &now)
	if !l.allow("1.1.1.1") {
		t.Fatal("allow(1.1.1.1 #1) = false, want true")
	}
	if l.allow("1.1.1.1") {
		t.Fatal("allow(1.1.1.1 #2) = true, want false")
	}
	if !l.allow("2.2.2.2") {
		t.Fatal("allow(2.2.2.2) = false, want true")
	}
}

func TestLimiterWindowReset(t *testing.T) {
	now := time.Now()
	l := newFakeLimiter(1, time.Minute, &now)
	if !l.allow("1.2.3.4") {
		t.Fatal("allow #1 = false, want true")
	}
	if l.allow("1.2.3.4") {
		t.Fatal("allow #2 = true, want false")
	}
	// 窗口前进 61 秒，应重置放行。
	now = now.Add(61 * time.Second)
	if !l.allow("1.2.3.4") {
		t.Fatal("allow after window reset = false, want true")
	}
}

func TestClientKey(t *testing.T) {
	// RemoteAddr 无 X-Forwarded-For 时取 host 部分。
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "127.0.0.1:8080"
	if got := clientKey(r); got != "127.0.0.1" {
		t.Fatalf("clientKey() = %q, want 127.0.0.1", got)
	}
	// X-Forwarded-For 优先。
	r.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := clientKey(r); got != "203.0.113.9" {
		t.Fatalf("clientKey() = %q, want 203.0.113.9", got)
	}
}

package middleware

import (
	"net/http"
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

func TestLimiterRemovesExpiredBuckets(t *testing.T) {
	now := time.Now()
	l := newFakeLimiter(1, time.Minute, &now)

	if !l.allow("1.1.1.1") {
		t.Fatal("allow(1.1.1.1) = false, want true")
	}
	if !l.allow("2.2.2.2") {
		t.Fatal("allow(2.2.2.2) = false, want true")
	}
	if got := len(l.buckets); got != 2 {
		t.Fatalf("len(buckets) = %d, want 2", got)
	}

	now = now.Add(61 * time.Second)
	if !l.allow("3.3.3.3") {
		t.Fatal("allow(3.3.3.3) = false, want true")
	}
	if got := len(l.buckets); got != 1 {
		t.Fatalf("len(buckets) = %d, want only current bucket after cleanup", got)
	}
	if _, ok := l.buckets["3.3.3.3"]; !ok {
		t.Fatal("current bucket was removed during cleanup")
	}
}

func TestClientKey(t *testing.T) {
	t.Setenv("DRIVER_TRUSTED_PROXY_HOPS", "")

	// RemoteAddr 无 X-Forwarded-For 时取 host 部分。
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "127.0.0.1:8080"
	if got := clientKey(r); got != "127.0.0.1" {
		t.Fatalf("clientKey() = %q, want 127.0.0.1", got)
	}
	// X-Forwarded-For is ignored unless trusted proxy hops are configured.
	r.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := clientKey(r); got != "127.0.0.1" {
		t.Fatalf("clientKey() = %q, want remote address when trusted proxy hops is not configured", got)
	}
}

func TestClientKeyUsesTrustedProxyHopFromRight(t *testing.T) {
	t.Setenv("DRIVER_TRUSTED_PROXY_HOPS", "1")

	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "10.0.0.10:8080"
	r.Header.Set("X-Forwarded-For", "198.51.100.99, 203.0.113.9")

	if got := clientKey(r); got != "203.0.113.9" {
		t.Fatalf("clientKey() = %q, want rightmost trusted proxy client", got)
	}
}

func TestClientKeyFallsBackWhenTrustedProxyHopIsInvalid(t *testing.T) {
	t.Setenv("DRIVER_TRUSTED_PROXY_HOPS", "2")

	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "10.0.0.10:8080"
	r.Header.Set("X-Forwarded-For", "203.0.113.9")

	if got := clientKey(r); got != "10.0.0.10" {
		t.Fatalf("clientKey() = %q, want remote address when X-Forwarded-For has too few hops", got)
	}
}

func TestLoginRateLimitIgnoresSpoofedXForwardedForByDefault(t *testing.T) {
	t.Setenv("DRIVER_TRUSTED_PROXY_HOPS", "")

	handler := LoginRateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRequest("POST", "/", nil)
	first.RemoteAddr = "192.0.2.10:12345"
	first.Header.Set("X-Forwarded-For", "198.51.100.1")
	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	if firstResponse.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", firstResponse.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest("POST", "/", nil)
	second.RemoteAddr = "192.0.2.10:12345"
	second.Header.Set("X-Forwarded-For", "198.51.100.2")
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)
	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondResponse.Code, http.StatusTooManyRequests)
	}
}

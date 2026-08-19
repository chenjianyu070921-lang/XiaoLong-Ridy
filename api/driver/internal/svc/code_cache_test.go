package svc

import (
	"testing"
	"time"
)

func TestCodeCacheSetAndVerify(t *testing.T) {
	c := NewCodeCache(time.Minute)
	c.Set("13800000001", "123456")
	if !c.Verify("13800000001", "123456") {
		t.Fatal("Verify(正确验证码) = false, want true")
	}
}

func TestCodeCacheSingleUse(t *testing.T) {
	c := NewCodeCache(time.Minute)
	c.Set("13800000001", "123456")
	if !c.Verify("13800000001", "123456") {
		t.Fatal("Verify #1 = false, want true")
	}
	// 一次性：再次使用同验证码必须失败（防重放）。
	if c.Verify("13800000001", "123456") {
		t.Fatal("Verify #2 = true, want false")
	}
}

func TestCodeCacheWrongCode(t *testing.T) {
	c := NewCodeCache(time.Minute)
	c.Set("13800000001", "123456")
	if c.Verify("13800000001", "000000") {
		t.Fatal("Verify(错误验证码) = true, want false")
	}
}

func TestCodeCacheNotFound(t *testing.T) {
	c := NewCodeCache(time.Minute)
	if c.Verify("13800000009", "123456") {
		t.Fatal("Verify(未发送验证码的手机号) = true, want false")
	}
}

func TestCodeCacheExpired(t *testing.T) {
	c := NewCodeCache(10 * time.Millisecond)
	c.Set("13800000001", "123456")
	time.Sleep(20 * time.Millisecond)
	if c.Verify("13800000001", "123456") {
		t.Fatal("Verify(过期验证码) = true, want false")
	}
}

func TestCodeCacheTTL(t *testing.T) {
	if got := NewCodeCache(5 * time.Minute).TTL(); got != 5*time.Minute {
		t.Fatalf("TTL() = %v, want 5m", got)
	}
}

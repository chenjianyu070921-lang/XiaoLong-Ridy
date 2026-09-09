package redisx

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

// TestTryLockWithMaxHold_StopsRenewing 验证锁超过 maxHold 后看门狗自动停止续期。
func TestTryLockWithMaxHold_StopsRenewing(t *testing.T) {
	client, mr := newTestRedis(t)
	defer client.Close()
	defer mr.Close()

	ttl := 200 * time.Millisecond
	maxHold := 400 * time.Millisecond

	lk, err := TryLockWithMaxHold(context.Background(), client, "test:lock:maxhold", ttl, maxHold)
	if err != nil {
		t.Fatalf("TryLockWithMaxHold failed: %v", err)
	}
	defer lk.Release()

	// 等待 maxHold + 安全余量，确保看门狗 ticker 已触发并检测到超时
	time.Sleep(600 * time.Millisecond)

	// 看门狗应已停止：done channel 已关闭
	select {
	case <-lk.done:
		// 符合预期
	default:
		t.Fatal("watchdog should have stopped after maxHold exceeded")
	}

	// 推进 miniredis 时钟超过 TTL，触发 key 过期
	mr.FastForward(ttl + 100*time.Millisecond)

	// Redis 中的锁 key 应已过期（看门狗停止续期后 TTL 自然到期）
	exists, err := client.Exists(context.Background(), "test:lock:maxhold").Result()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists != 0 {
		t.Fatal("lock key should have expired after watchdog stopped renewing")
	}
}

// TestTryLockWithoutRelease_Expires 验证锁未 Release 时不会永久占用，
// 超过 maxHold 后看门狗停止续期，TTL 到期后锁自动释放。
func TestTryLockWithoutRelease_Expires(t *testing.T) {
	client, mr := newTestRedis(t)
	defer client.Close()
	defer mr.Close()

	ttl := 200 * time.Millisecond
	maxHold := 400 * time.Millisecond
	key := "test:lock:norelease"

	lk, err := TryLockWithMaxHold(context.Background(), client, key, ttl, maxHold)
	if err != nil {
		t.Fatalf("TryLockWithMaxHold failed: %v", err)
	}
	// 故意不调用 Release，模拟代码漏调场景

	// 等待 maxHold，确保看门狗检测到超时并停止续期
	time.Sleep(600 * time.Millisecond)

	// 推进 miniredis 时钟超过 TTL，触发 key 过期
	mr.FastForward(ttl + 100*time.Millisecond)

	// 锁应已过期：新请求可以获取同一 key 的锁
	lk2, err := TryLock(context.Background(), client, key, ttl)
	if err != nil {
		t.Fatalf("should be able to acquire lock after expiry without Release, got: %v", err)
	}
	defer lk2.Release()

	// 确认 lk 和 lk2 不是同一把锁（token 不同）
	if lk.token == lk2.token {
		t.Fatal("new lock should have a different token")
	}
}

// TestTryLock_BackwardCompatibility 验证 TryLock（无 maxHold）保持向后兼容，
// 看门狗持续续期直到 Release 被调用。
func TestTryLock_BackwardCompatibility(t *testing.T) {
	client, mr := newTestRedis(t)
	defer client.Close()
	defer mr.Close()

	ttl := 200 * time.Millisecond

	lk, err := TryLock(context.Background(), client, "test:lock:compat", ttl)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// 等待超过 TTL，看门狗应已续期，锁仍然存在
	time.Sleep(350 * time.Millisecond)

	exists, err := client.Exists(context.Background(), "test:lock:compat").Result()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists != 1 {
		t.Fatal("lock should still exist with watchdog renewing (no maxHold)")
	}

	lk.Release()

	// Release 后锁应被删除
	exists, err = client.Exists(context.Background(), "test:lock:compat").Result()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists != 0 {
		t.Fatal("lock should be deleted after Release")
	}
}

// TestTryLock_ErrLockHeld 验证锁被占用时返回 ErrLockHeld。
func TestTryLock_ErrLockHeld(t *testing.T) {
	client, mr := newTestRedis(t)
	defer client.Close()
	defer mr.Close()

	key := "test:lock:held"
	lk, err := TryLock(context.Background(), client, key, 5*time.Second)
	if err != nil {
		t.Fatalf("first TryLock failed: %v", err)
	}
	defer lk.Release()

	_, err = TryLock(context.Background(), client, key, 5*time.Second)
	if err != ErrLockHeld {
		t.Fatalf("second TryLock should return ErrLockHeld, got: %v", err)
	}
}

// TestRelease_Idempotent 验证 Release 可重复调用不 panic。
func TestRelease_Idempotent(t *testing.T) {
	client, mr := newTestRedis(t)
	defer client.Close()
	defer mr.Close()

	lk, err := TryLock(context.Background(), client, "test:lock:idempotent", 5*time.Second)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	lk.Release()
	lk.Release() // 不应 panic
	lk.Release() // 不应 panic
}

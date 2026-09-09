package onlinestore

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestStore 启动内存 Redis 并返回在线状态 Store 与清理函数。
func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s := NewStore(rdb, time.Second)
	return s, func() { _ = rdb.Close(); mr.Close() }
}

// TestSetOnlineAndGet 验证上线写入后能读到在线状态与设备标识。
func TestSetOnlineAndGet(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()

	if err := s.SetOnline(context.Background(), 100, "dev-A", 116.40, 39.90); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}
	st, err := s.Get(context.Background(), 100)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if st == nil {
		t.Fatal("expected state, got nil")
	}
	if st.OnlineStatus != Online {
		t.Fatalf("expected online, got %d", st.OnlineStatus)
	}
	if st.DeviceID != "dev-A" {
		t.Fatalf("expected device dev-A, got %s", st.DeviceID)
	}
}

// TestHeartbeatSameDevice 验证同一设备心跳不触发互踢，且保持在线。
func TestHeartbeatSameDevice(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.SetOnline(ctx, 100, "dev-A", 116.40, 39.90); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}
	status, kicked, err := s.Heartbeat(ctx, 100, "dev-A", 116.41, 39.91)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
	if kicked {
		t.Fatal("same device should not be kicked")
	}
	if status != Online {
		t.Fatalf("expected online, got %d", status)
	}
}

// TestHeartbeatOtherDevice 验证不同设备心跳判定被顶替（kicked=true），且不覆盖在线状态。
func TestHeartbeatOtherDevice(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.SetOnline(ctx, 100, "dev-A", 116.40, 39.90); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}
	_, kicked, err := s.Heartbeat(ctx, 100, "dev-B", 116.41, 39.91)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
	if !kicked {
		t.Fatal("other device should be kicked")
	}
	// 被踢设备心跳不应改写已绑定设备，保证 dev-A 仍是当前设备。
	st, err := s.Get(ctx, 100)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if st.DeviceID != "dev-A" {
		t.Fatalf("expected binding dev-A preserved, got %s", st.DeviceID)
	}
}

// TestSetOffline 验证下线后状态变为离线。
func TestSetOffline(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.SetOnline(ctx, 100, "dev-A", 116.40, 39.90); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}
	if err := s.SetOffline(ctx, 100); err != nil {
		t.Fatalf("SetOffline failed: %v", err)
	}
	st, err := s.Get(ctx, 100)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if st.OnlineStatus != Offline {
		t.Fatalf("expected offline, got %d", st.OnlineStatus)
	}
}

// TestHeartbeatNoRecord 验证无在线记录时心跳会建立在线状态且不触发互踢。
func TestHeartbeatNoRecord(t *testing.T) {
	s, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	status, kicked, err := s.Heartbeat(ctx, 200, "dev-C", 0, 0)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
	if kicked {
		t.Fatal("new device should not be kicked")
	}
	if status != Online {
		t.Fatalf("expected online, got %d", status)
	}
}

func TestDefaultTTLSurvivesOneMobileTimerClamp(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	defer mr.Close()

	s := NewStore(rdb, 0)
	ctx := context.Background()
	if err := s.SetOnline(ctx, 300, "dev-mobile", 116.40, 39.90); err != nil {
		t.Fatalf("SetOnline failed: %v", err)
	}

	mr.FastForward(65 * time.Second)

	st, err := s.Get(ctx, 300)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if st == nil || st.OnlineStatus != Online {
		t.Fatalf("expected driver to remain online after one clamped heartbeat window, got %+v", st)
	}
}

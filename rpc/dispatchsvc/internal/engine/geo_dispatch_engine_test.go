package engine

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

// TestGeoDispatchEngineMockFallbackDisabled 验证默认（EnableMockDispatch=false）
// 时 GEO 查不到司机不会回退 mock：返回空候选而非假成功。
func TestGeoDispatchEngineMockFallbackDisabled(t *testing.T) {
	eng, ok := NewGeoDispatchEngine(nil, "default").(*geoDispatchEngine)
	if !ok {
		t.Fatalf("NewGeoDispatchEngine() type = %T, want *geoDispatchEngine", eng)
	}
	if eng.enableMock {
		t.Fatal("NewGeoDispatchEngine() should default to enableMock=false")
	}
	candidates, err := eng.FindCandidates(context.Background(), 1, 116.47, 39.9, 1, "110000", OrderTypeRealtime)
	if err != nil {
		t.Fatalf("FindCandidates() error = %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("FindCandidates() len = %d, want 0 when mock fallback disabled", len(candidates))
	}
}

// TestGeoDispatchEngineMockFallbackEnabled 验证显式开启 EnableMockDispatch
// 时 GEO 无数据会回退 mock 候选（联调场景）。
func TestGeoDispatchEngineMockFallbackEnabled(t *testing.T) {
	eng, ok := NewGeoDispatchEngineWithMock(nil, "default", true).(*geoDispatchEngine)
	if !ok {
		t.Fatalf("NewGeoDispatchEngineWithMock() type = %T, want *geoDispatchEngine", eng)
	}
	if !eng.enableMock {
		t.Fatal("NewGeoDispatchEngineWithMock(true) should set enableMock=true")
	}
	candidates, err := eng.FindCandidates(context.Background(), 1, 116.47, 39.9, 1, "110000", OrderTypeRealtime)
	if err != nil {
		t.Fatalf("FindCandidates() error = %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("FindCandidates() should fallback to mock candidates when enableMock=true")
	}
}

// TestGeoDispatchEngineAvailabilityFilter 验证可用性过滤：
// 只保留"在线且未忙碌"的司机（P1-M4-8）。
func TestGeoDispatchEngineAvailabilityFilter(t *testing.T) {
	eng := &geoDispatchEngine{
		availability: func(_ context.Context, driverID uint64) (online, busy bool) {
			switch driverID {
			case 1: // 在线且空闲 → 保留
				return true, false
			case 2: // 在线但忙碌 → 剔除
				return true, true
			case 3: // 离线 → 剔除
				return false, false
			default:
				return true, false
			}
		},
	}
	locs := []redis.GeoLocation{
		{Name: "1"}, {Name: "2"}, {Name: "3"},
	}
	filtered := eng.filterAvailable(context.Background(), locs)
	if len(filtered) != 1 || filtered[0].Name != "1" {
		t.Fatalf("filterAvailable() = %+v, want only driver 1", filtered)
	}
}

// TestGeoDispatchEngineAvailabilityNil 验证未注入 availability 时原样返回（兼容旧行为）。
func TestGeoDispatchEngineAvailabilityNil(t *testing.T) {
	eng := &geoDispatchEngine{}
	locs := []redis.GeoLocation{{Name: "1"}, {Name: "2"}}
	filtered := eng.filterAvailable(context.Background(), locs)
	if len(filtered) != 2 {
		t.Fatalf("filterAvailable() with nil availability len = %d, want 2", len(filtered))
	}
}

func TestGeoDispatchEnginePreferenceFilter(t *testing.T) {
	eng := &geoDispatchEngine{
		preference: func(_ context.Context, driverID uint64, orderType int32) bool {
			return driverID == 1 && orderType == OrderTypeReservation
		},
	}
	locs := []redis.GeoLocation{{Name: "1"}, {Name: "2"}}
	filtered := eng.filterPreference(context.Background(), locs, OrderTypeReservation)
	if len(filtered) != 1 || filtered[0].Name != "1" {
		t.Fatalf("filterPreference() = %+v, want only driver 1", filtered)
	}
}

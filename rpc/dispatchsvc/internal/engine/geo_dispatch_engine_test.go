package engine

import (
	"context"
	"testing"
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
	candidates, err := eng.FindCandidates(context.Background(), 1, 116.47, 39.9, 1, "110000")
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
	candidates, err := eng.FindCandidates(context.Background(), 1, 116.47, 39.9, 1, "110000")
	if err != nil {
		t.Fatalf("FindCandidates() error = %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("FindCandidates() should fallback to mock candidates when enableMock=true")
	}
}

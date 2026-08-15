package engine

import (
	"context"
	"math"
)

// Candidate 表示派单候选司机。
type Candidate struct {
	DriverID   uint64
	MatchScore float64
}

// DispatchEngine 定义候选司机匹配接口。
type DispatchEngine interface {
	FindCandidates(ctx context.Context, orderID uint64, fromLongitude, fromLatitude float64, carType int32, cityCode string) ([]Candidate, error)
}

type mockDispatchEngine struct {
	driverIDs []uint64
}

// NewMockDispatchEngine 创建 P0 联调用的 mock 派单引擎。
func NewMockDispatchEngine() DispatchEngine {
	return &mockDispatchEngine{
		driverIDs: []uint64{9001, 9002, 9003},
	}
}

// FindCandidates 返回固定候选司机，并用 mock 距离计算匹配分。
func (e *mockDispatchEngine) FindCandidates(_ context.Context, _ uint64, fromLongitude, fromLatitude float64, _ int32, _ string) ([]Candidate, error) {
	candidates := make([]Candidate, 0, len(e.driverIDs))
	for i, driverID := range e.driverIDs {
		// mock 距离：以起点为中心按索引错开 0.8km，仅用于生成可排序的匹配分。
		distanceKm := 1.0 + float64(i)*0.8
		score := math.Max(0, 100-distanceKm*5)
		_ = fromLongitude
		_ = fromLatitude
		candidates = append(candidates, Candidate{DriverID: driverID, MatchScore: score})
	}
	return candidates, nil
}

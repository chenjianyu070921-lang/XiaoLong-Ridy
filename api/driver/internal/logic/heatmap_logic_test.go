package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type fakeHeatmapRepository struct {
	calls      int
	lastRadius float64
	rows       []svc.HeatmapOrderLocation
}

func (f *fakeHeatmapRepository) ListWaitAcceptOrderLocations(_ context.Context, longitude, latitude, radiusMeters float64) ([]svc.HeatmapOrderLocation, error) {
	f.calls++
	f.lastRadius = radiusMeters
	return f.rows, nil
}

func TestGetOrderHeatmapAggregatesAndCaches(t *testing.T) {
	repo := &fakeHeatmapRepository{
		rows: []svc.HeatmapOrderLocation{
			{Longitude: 116.397000, Latitude: 39.908000},
			{Longitude: 116.397500, Latitude: 39.908500},
			{Longitude: 116.410000, Latitude: 39.908000},
		},
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	logic := NewHeatmapLogic(context.Background(), &svc.ServiceContext{HeatmapRepository: repo, RedisClient: rdb})
	req := &types.HeatmapRequest{
		Longitude:    116.397,
		Latitude:     39.908,
		RadiusMeters: 2000,
	}

	resp, err := logic.GetOrderHeatmap(25, req)
	if err != nil {
		t.Fatalf("GetOrderHeatmap() error = %v", err)
	}
	if resp.GridSizeMeters != 200 || resp.RadiusMeters != 2000 {
		t.Fatalf("GetOrderHeatmap() metadata = %+v", resp)
	}
	if len(resp.Points) != 2 {
		t.Fatalf("GetOrderHeatmap() points = %+v, want 2 aggregated grid points", resp.Points)
	}
	if resp.Points[0].Weight != 2 || resp.Points[1].Weight != 1 {
		t.Fatalf("GetOrderHeatmap() weights = %+v, want [2 1]", resp.Points)
	}

	second, err := logic.GetOrderHeatmap(25, req)
	if err != nil {
		t.Fatalf("second GetOrderHeatmap() error = %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repository calls = %d, want 1 due to short cache", repo.calls)
	}
	if len(second.Points) != len(resp.Points) || second.Points[0].Weight != resp.Points[0].Weight {
		t.Fatalf("cached response = %+v, want %+v", second, resp)
	}
}

func TestGetOrderHeatmapRejectsInvalidLocationAndRadius(t *testing.T) {
	logic := NewHeatmapLogic(context.Background(), &svc.ServiceContext{HeatmapRepository: &fakeHeatmapRepository{}})

	tests := []*types.HeatmapRequest{
		nil,
		{Longitude: 0, Latitude: 39.908, RadiusMeters: 1000},
		{Longitude: 116.397, Latitude: 0, RadiusMeters: 1000},
		{Longitude: 116.397, Latitude: 39.908, RadiusMeters: 0},
		{Longitude: 116.397, Latitude: 39.908, RadiusMeters: -1},
	}
	for _, req := range tests {
		if _, err := logic.GetOrderHeatmap(25, req); err != ErrInvalidParam {
			t.Fatalf("GetOrderHeatmap(%+v) error = %v, want %v", req, err, ErrInvalidParam)
		}
	}
}

func TestGetOrderHeatmapClampsRadiusToFiveKilometers(t *testing.T) {
	repo := &fakeHeatmapRepository{}
	logic := NewHeatmapLogic(context.Background(), &svc.ServiceContext{HeatmapRepository: repo})

	resp, err := logic.GetOrderHeatmap(25, &types.HeatmapRequest{
		Longitude:    116.397,
		Latitude:     39.908,
		RadiusMeters: 8000,
	})
	if err != nil {
		t.Fatalf("GetOrderHeatmap() error = %v", err)
	}
	if resp.RadiusMeters != 5000 {
		t.Fatalf("response radius = %v, want 5000", resp.RadiusMeters)
	}
	if repo.lastRadius != 5000 {
		t.Fatalf("repository radius = %v, want 5000", repo.lastRadius)
	}
}

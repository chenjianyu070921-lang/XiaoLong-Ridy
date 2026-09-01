package logic

import (
	"context"
	"fmt"
	"math"
	"sort"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

const (
	heatmapGridSizeMeters  = 200.0
	maxHeatmapRadiusMeters = 5000.0
)

type HeatmapLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHeatmapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HeatmapLogic {
	return &HeatmapLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *HeatmapLogic) GetOrderHeatmap(driverID int64, req *types.HeatmapRequest) (*types.HeatmapResponse, error) {
	if driverID <= 0 || req == nil || req.Longitude == 0 || req.Latitude == 0 || !validLocation(req.Longitude, req.Latitude) || req.RadiusMeters <= 0 {
		return nil, ErrInvalidParam
	}
	radius := req.RadiusMeters
	if radius > maxHeatmapRadiusMeters {
		radius = maxHeatmapRadiusMeters
	}
	if l.svcCtx == nil || l.svcCtx.HeatmapRepository == nil {
		return nil, ErrHeatmapStorageNotConfigured
	}
	locations, err := l.svcCtx.HeatmapRepository.ListWaitAcceptOrderLocations(l.ctx, req.Longitude, req.Latitude, radius)
	if err != nil {
		return nil, err
	}
	resp := &types.HeatmapResponse{
		Points:         aggregateHeatmapPoints(req.Longitude, req.Latitude, radius, locations),
		RadiusMeters:   radius,
		GridSizeMeters: int64(heatmapGridSizeMeters),
	}
	return resp, nil
}

func aggregateHeatmapPoints(centerLon, centerLat, radiusMeters float64, locations []svc.HeatmapOrderLocation) []types.HeatmapPoint {
	type bucket struct {
		lonSum float64
		latSum float64
		count  int64
	}
	buckets := make(map[string]*bucket)
	for _, location := range locations {
		if !validLocation(location.Longitude, location.Latitude) {
			continue
		}
		if haversineMeters(centerLon, centerLat, location.Longitude, location.Latitude) > radiusMeters {
			continue
		}
		x, y := heatmapCell(location.Longitude, location.Latitude)
		key := fmt.Sprintf("%d:%d", x, y)
		cell := buckets[key]
		if cell == nil {
			cell = &bucket{}
			buckets[key] = cell
		}
		cell.lonSum += location.Longitude
		cell.latSum += location.Latitude
		cell.count++
	}
	points := make([]types.HeatmapPoint, 0, len(buckets))
	for _, cell := range buckets {
		points = append(points, types.HeatmapPoint{
			Longitude: roundCoord(cell.lonSum / float64(cell.count)),
			Latitude:  roundCoord(cell.latSum / float64(cell.count)),
			Weight:    cell.count,
		})
	}
	sort.SliceStable(points, func(i, j int) bool {
		if points[i].Weight == points[j].Weight {
			if points[i].Longitude == points[j].Longitude {
				return points[i].Latitude < points[j].Latitude
			}
			return points[i].Longitude < points[j].Longitude
		}
		return points[i].Weight > points[j].Weight
	})
	return points
}

func heatmapCell(longitude, latitude float64) (int64, int64) {
	xMeters := longitude * 111320.0 * math.Cos(latitude*math.Pi/180)
	yMeters := latitude * 110540.0
	return int64(math.Floor(xMeters / heatmapGridSizeMeters)), int64(math.Floor(yMeters / heatmapGridSizeMeters))
}

func roundCoord(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}

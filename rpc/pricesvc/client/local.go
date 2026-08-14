package client

import (
	"context"
	"math"
)

// EstimatePriceRequest 表示预估价格所需的核心参数。
type EstimatePriceRequest struct {
	CarType         int32
	FromLongitude   float64
	FromLatitude    float64
	ToLongitude     float64
	ToLatitude      float64
	EstimatedMeters int64
	EstimatedSecond int64
}

// EstimatePriceResponse 返回预估价格、里程和时长。
type EstimatePriceResponse struct {
	EstimatedPriceCents int64
	EstimatedDistanceM  int64
	EstimatedDurationS  int64
}

// LocalClient 是本地开发和测试使用的价格服务实现。
type LocalClient struct {
}

// NewLocalClient 创建本地价格服务实现。
func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

// EstimatePrice 根据起终点坐标计算一个稳定可复现的预估价格。
func (c *LocalClient) EstimatePrice(_ context.Context, req *EstimatePriceRequest) (*EstimatePriceResponse, error) {
	distanceM := req.EstimatedMeters
	if distanceM <= 0 {
		distanceM = haversineMeters(req.FromLatitude, req.FromLongitude, req.ToLatitude, req.ToLongitude)
	}
	if distanceM <= 0 {
		distanceM = 1000
	}

	durationS := req.EstimatedSecond
	if durationS <= 0 {
		durationS = int64(math.Ceil(float64(distanceM) / 250.0))
	}
	if durationS <= 0 {
		durationS = 60
	}

	base := int64(300)
	perKm := int64(180)
	switch req.CarType {
	case 1:
		base = 200
		perKm = 140
	case 2:
		base = 300
		perKm = 180
	case 3:
		base = 500
		perKm = 260
	}
	priceCents := base + int64(math.Ceil(float64(distanceM)/1000.0))*perKm + int64(math.Ceil(float64(durationS)/60.0))*10
	if priceCents < 0 {
		priceCents = 0
	}
	return &EstimatePriceResponse{
		EstimatedPriceCents: priceCents,
		EstimatedDistanceM:  distanceM,
		EstimatedDurationS:  durationS,
	}, nil
}

// haversineMeters 使用球面距离公式估算两组经纬度之间的直线距离。
func haversineMeters(lat1, lon1, lat2, lon2 float64) int64 {
	const earthRadius = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return int64(math.Round(earthRadius * c))
}

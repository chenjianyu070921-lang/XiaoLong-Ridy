package logic

import (
	"context"
	"math"

	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoutePlanLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoutePlanLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoutePlanLogic {
	return &RoutePlanLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RoutePlan 路径规划（计算距离和预计时间）
func (l *RoutePlanLogic) RoutePlan(in *locationsvc.RoutePlanReq) (*locationsvc.RoutePlanResp, error) {
	l.Infof("RoutePlan: origin(%f,%f) -> dest(%f,%f)",
		in.OriginLat, in.OriginLng, in.DestinationLat, in.DestinationLng)

	distance := haversine(in.OriginLat, in.OriginLng, in.DestinationLat, in.DestinationLng)
	duration := int32(float64(distance) / 30000.0 * 3600)

	return &locationsvc.RoutePlanResp{
		Distance:   int32(distance),
		Duration:   duration,
		Polyline:   "",
		OriginLat:  in.OriginLat,
		OriginLng:  in.OriginLng,
		DestLat:    in.DestinationLat,
		DestLng:    in.DestinationLng,
	}, nil
}

// haversine 计算两点间球面距离（米）
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

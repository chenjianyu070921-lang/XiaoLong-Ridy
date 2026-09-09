package logic

import (
	"context"

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

// RoutePlan 路径规划，调用高德 direction/driving 接口返回真实可行驶路线
func (l *RoutePlanLogic) RoutePlan(in *locationsvc.RoutePlanReq) (*locationsvc.RoutePlanResp, error) {
	l.Infof("RoutePlan: origin(%f,%f) -> dest(%f,%f)",
		in.OriginLat, in.OriginLng, in.DestinationLat, in.DestinationLng)

	route, err := l.svcCtx.GetGeo().RoutePlan(in.OriginLat, in.OriginLng, in.DestinationLat, in.DestinationLng)
	if err != nil {
		l.Errorf("调用高德驾车路径规划失败: %v", err)
		return nil, err
	}

	return &locationsvc.RoutePlanResp{
		Distance:   route.Distance(),
		Duration:   route.Duration(),
		Polyline:   route.Polyline(),
		OriginLat:  in.OriginLat,
		OriginLng:  in.OriginLng,
		DestLat:    in.DestinationLat,
		DestLng:    in.DestinationLng,
	}, nil
}

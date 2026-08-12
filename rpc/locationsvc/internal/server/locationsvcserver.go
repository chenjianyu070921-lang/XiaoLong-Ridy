package server

import (
	"context"

	"XiaoLong-Ridy/rpc/locationsvc/internal/logic"
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	"XiaoLong-Ridy/rpc/locationsvc/locationsvc"
)

type LocationServiceServer struct {
	svcCtx *svc.ServiceContext
	locationsvc.UnimplementedLocationServiceServer
}

func NewLocationServiceServer(svcCtx *svc.ServiceContext) *LocationServiceServer {
	return &LocationServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *LocationServiceServer) ReverseGeocode(ctx context.Context, in *locationsvc.ReverseGeocodeReq) (*locationsvc.ReverseGeocodeResp, error) {
	l := logic.NewReverseGeocodeLogic(ctx, s.svcCtx)
	return l.ReverseGeocode(in)
}

func (s *LocationServiceServer) POISearch(ctx context.Context, in *locationsvc.POISearchReq) (*locationsvc.POISearchResp, error) {
	l := logic.NewPOISearchLogic(ctx, s.svcCtx)
	return l.POISearch(in)
}

func (s *LocationServiceServer) RoutePlan(ctx context.Context, in *locationsvc.RoutePlanReq) (*locationsvc.RoutePlanResp, error) {
	l := logic.NewRoutePlanLogic(ctx, s.svcCtx)
	return l.RoutePlan(in)
}

package logic

import (
	"XiaoLong-Ridy/rpc/locationsvc/internal/svc"
	locationsvc "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
	"context"
)

// GetDriverLocationLogic 负责读取司机最新上报位置，向乘客提供实时追踪数据。
type GetDriverLocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetDriverLocationLogic 创建司机位置查询逻辑对象。
func NewGetDriverLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLocationLogic {
	return &GetDriverLocationLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetDriverLocation 按司机 ID 查询最新位置及在线状态。
func (l *GetDriverLocationLogic) GetDriverLocation(in *locationsvc.GetDriverLocationReq) (*locationsvc.GetDriverLocationResp, error) {
	loc, err := l.svcCtx.DriverLocationModel.GetByDriverID(uint64(in.GetDriverId()))
	if err != nil {
		return nil, err
	}
	return &locationsvc.GetDriverLocationResp{DriverId: int64(loc.DriverID), Lng: loc.Longitude, Lat: loc.Latitude, Heading: int32(loc.Heading), SpeedKmh: loc.SpeedKmh, OnlineStatus: int32(loc.OnlineStatus), ReportTime: loc.ReportTime.Unix()}, nil
}

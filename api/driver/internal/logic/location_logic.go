package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type LocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LocationLogic {
	return &LocationLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *LocationLogic) ReportLocation(driverID int64, req *types.ReportLocationRequest) (*types.ReportLocationResponse, error) {
	if driverID <= 0 || req == nil || strings.TrimSpace(req.DeviceID) == "" ||
		!validLocation(req.Longitude, req.Latitude) {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ReportLocation(l.ctx, &driversproto.ReportLocationRequest{
		DriverId:  driverID,
		DeviceId:  strings.TrimSpace(req.DeviceID),
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	})
	if err != nil {
		return nil, err
	}
	if req.OrderID > 0 {
		if l.svcCtx != nil && l.svcCtx.TrajectoryRepository != nil {
			if trajErr := l.svcCtx.TrajectoryRepository.RecordPoint(l.ctx, &svc.TrajectoryRecord{
				OrderID:   req.OrderID,
				DriverID:  driverID,
				Longitude: req.Longitude,
				Latitude:  req.Latitude,
				SpeedKmh:  req.SpeedKmh,
				Heading:   req.Heading,
			}); trajErr != nil {
				logx.WithContext(l.ctx).Errorf("trajectory record failed (driver location already updated): %v", trajErr)
			}
		}
	}
	return &types.ReportLocationResponse{
		DriverID:     resp.GetDriverId(),
		OnlineStatus: int(resp.GetOnlineStatus()),
		Kicked:       resp.GetKicked(),
		ReportTime:   resp.GetReportTime(),
	}, nil
}

func (l *LocationLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

var errInvalidLongitudeLatitude = errors.New("经纬度不合法")

type ReportLocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportLocationLogic {
	return &ReportLocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReportLocation 刷新司机在线保活并同步最新位置到 driver_location。
func (l *ReportLocationLogic) ReportLocation(in *proto.ReportLocationRequest) (*proto.ReportLocationResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if in.GetDeviceId() == "" {
		return nil, errInvalidDeviceID
	}
	if !validLongitudeLatitude(in.GetLongitude(), in.GetLatitude()) {
		return nil, errInvalidLongitudeLatitude
	}
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}

	onlineStatus, kicked, err := l.svcCtx.OnlineStore.Heartbeat(l.ctx, in.GetDriverId(), in.GetDeviceId(), in.GetLongitude(), in.GetLatitude())
	if err != nil {
		return nil, err
	}
	reportTime := time.Now()
	if !kicked {
		if err := l.svcCtx.DriverRepository.UpsertLocation(l.ctx, &model.DriverLocation{
			DriverID:     uint64(in.GetDriverId()),
			Longitude:    in.GetLongitude(),
			Latitude:     in.GetLatitude(),
			OnlineStatus: locationStatusFromOnline(onlineStatus),
			ReportTime:   reportTime,
		}); err != nil {
			return nil, err
		}
		if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.GetDriverId()), map[string]interface{}{"online_status": locationStatusFromOnline(onlineStatus)}); err != nil {
			return nil, err
		}
		if onlineStatus == int32(DriverOnline) {
			if err := syncDispatchDriverOnline(l.ctx, l.svcCtx, in.GetDriverId(), in.GetLongitude(), in.GetLatitude()); err != nil {
				return nil, err
			}
		} else if onlineStatus == int32(DriverOffline) {
			if err := syncDispatchDriverOffline(l.ctx, l.svcCtx, in.GetDriverId()); err != nil {
				return nil, err
			}
		}
	}
	return &proto.ReportLocationResponse{
		DriverId:     in.GetDriverId(),
		OnlineStatus: onlineStatus,
		Kicked:       kicked,
		ReportTime:   reportTime.Unix(),
	}, nil
}

func validLongitudeLatitude(longitude, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90
}

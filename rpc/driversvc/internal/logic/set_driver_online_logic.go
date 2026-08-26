package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// DriverOnlineStatus 司机上下线状态值。
const (
	// DriverOffline 表示司机离线（不接单）。
	DriverOffline int8 = 0
	// DriverOnline 表示司机在线（可接单）。
	DriverOnline int8 = 1
)

type SetDriverOnlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDriverOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDriverOnlineLogic {
	return &SetDriverOnlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetDriverOnline 将司机听单状态置为在线（1）。
func (l *SetDriverOnlineLogic) SetDriverOnline(in *proto.SetDriverOnlineRequest) (*proto.SetDriverOnlineResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if in.GetDeviceId() == "" {
		return nil, errInvalidDeviceID
	}
	if !validLongitudeLatitude(in.GetLongitude(), in.GetLatitude()) {
		return nil, errInvalidLongitudeLatitude
	}
	// 先校验司机存在（软删不可见）。
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.DriverId)); err != nil {
		return nil, err
	}
	if err := l.svcCtx.OnlineStore.SetOnline(l.ctx, in.GetDriverId(), in.GetDeviceId(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	reportTime := time.Now()
	if err := l.svcCtx.DriverRepository.UpsertLocation(l.ctx, &model.DriverLocation{
		DriverID:     uint64(in.GetDriverId()),
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		OnlineStatus: DriverOnline,
		ReportTime:   reportTime,
	}); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"online_status": DriverOnline}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.DriverId), updates); err != nil {
		return nil, err
	}
	if err := syncDispatchDriverOnline(l.ctx, l.svcCtx, in.GetDriverId(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	return &proto.SetDriverOnlineResponse{
		DriverId:     in.DriverId,
		OnlineStatus: int32(DriverOnline),
	}, nil
}

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

// SetDriverOnline 将司机听单状态置为在线（1），并同步写入 Redis 在线状态与 MySQL 位置。
// 多端互踢：上线时在 Redis 绑定当前 device_id，供心跳/后续请求判定是否被新设备顶替。
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
	// 先判定是否已被其他设备顶替：若 Redis 中已有绑定且 device_id 不一致，则不覆盖并标记踢出。
	kicked := false
	if st, err := l.svcCtx.OnlineStore.Get(l.ctx, in.DriverId); err == nil && st != nil && st.DeviceID != "" && st.DeviceID != in.GetDeviceId() {
		kicked = true
		return &proto.SetDriverOnlineResponse{
			DriverId:     in.DriverId,
			OnlineStatus: int32(DriverOnline),
			Kicked:       kicked,
		}, nil
	}
	// 写入 Redis 在线状态并绑定设备，TTL 保活。
	if err := l.svcCtx.OnlineStore.SetOnline(l.ctx, in.DriverId, in.GetDeviceId(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	// 同步写入 MySQL 在线状态。
	updates := map[string]interface{}{"online_status": DriverOnline}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.DriverId), updates); err != nil {
		return nil, err
	}
	// 同步写入/更新司机最新位置到 driver_location，供附近司机查询使用。
	if err := l.svcCtx.DriverRepository.UpsertLocation(l.ctx, &model.DriverLocation{
		DriverID:     uint64(in.DriverId),
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		OnlineStatus: model.LocationOnline,
		ReportTime:   time.Now(),
	}); err != nil {
		return nil, err
	}
	return &proto.SetDriverOnlineResponse{
		DriverId:     in.DriverId,
		OnlineStatus: int32(DriverOnline),
		Kicked:       kicked,
	}, nil
}

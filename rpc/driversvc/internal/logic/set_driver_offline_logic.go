package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDriverOfflineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDriverOfflineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDriverOfflineLogic {
	return &SetDriverOfflineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetDriverOffline 将司机听单状态置为离线（0），并同步写入 Redis 离线状态与 MySQL 位置。
// 多端互踢：下线请求仅允许当前绑定设备执行，否则标记被顶替。
func (l *SetDriverOfflineLogic) SetDriverOffline(in *proto.SetDriverOfflineRequest) (*proto.SetDriverOfflineResponse, error) {
	// 先校验司机存在（软删不可见）。
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.DriverId)); err != nil {
		return nil, err
	}
	// 互踢判定：若 Redis 中绑定设备与请求设备不一致，则该设备已被顶替，不再执行下线。
	kicked := false
	if st, err := l.svcCtx.OnlineStore.Get(l.ctx, in.DriverId); err == nil && st != nil && st.DeviceID != "" && st.DeviceID != in.GetDeviceId() {
		kicked = true
		return &proto.SetDriverOfflineResponse{
			DriverId:     in.DriverId,
			OnlineStatus: int32(DriverOffline),
			Kicked:       kicked,
		}, nil
	}
	// 写入 Redis 离线状态。
	if err := l.svcCtx.OnlineStore.SetOffline(l.ctx, in.DriverId); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"online_status": DriverOffline}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.DriverId), updates); err != nil {
		return nil, err
	}
	// 同步更新司机位置表：置为离线状态，便于附近司机查询过滤。
	if err := l.svcCtx.DriverRepository.UpsertLocation(l.ctx, &model.DriverLocation{
		DriverID:     uint64(in.DriverId),
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		OnlineStatus: model.LocationOffline,
		ReportTime:   time.Now(),
	}); err != nil {
		return nil, err
	}
	return &proto.SetDriverOfflineResponse{
		DriverId:     in.DriverId,
		OnlineStatus: int32(DriverOffline),
		Kicked:       kicked,
	}, nil
}

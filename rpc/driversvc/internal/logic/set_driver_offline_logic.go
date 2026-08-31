package logic

import (
	"context"
	"errors"

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

// SetDriverOffline 将司机听单状态置为离线（0）。
func (l *SetDriverOfflineLogic) SetDriverOffline(in *proto.SetDriverOfflineRequest) (*proto.SetDriverOfflineResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil || l.svcCtx.OnlineStore == nil {
		return nil, errors.New("driver dependencies not ready")
	}
	// 先校验司机存在（软删不可见）。
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}
	if err := l.svcCtx.OnlineStore.SetOffline(l.ctx, in.GetDriverId()); err != nil {
		return nil, err
	}
	// 合并为一个事务，避免中间状态不一致
	if err := l.svcCtx.DriverRepository.UpdateStatusAndLocation(l.ctx, uint64(in.GetDriverId()), DriverOffline); err != nil {
		return nil, err
	}
	if err := syncDispatchDriverOffline(l.ctx, l.svcCtx, in.GetDriverId()); err != nil {
		return nil, err
	}
	return &proto.SetDriverOfflineResponse{
		DriverId:     in.GetDriverId(),
		OnlineStatus: int32(DriverOffline),
	}, nil
}

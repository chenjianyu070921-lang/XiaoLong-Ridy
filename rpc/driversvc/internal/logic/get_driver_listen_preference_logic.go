package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverListenPreferenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverListenPreferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverListenPreferenceLogic {
	return &GetDriverListenPreferenceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDriverListenPreferenceLogic) GetDriverListenPreference(in *proto.GetDriverListenPreferenceRequest) (*proto.DriverListenPreferenceResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	pref := defaultDriverListenPreference()
	var updatedAt int64
	if l.svcCtx != nil && l.svcCtx.DriverListenPreferenceRepository != nil {
		saved, err := l.svcCtx.DriverListenPreferenceRepository.GetByDriverID(l.ctx, uint64(in.GetDriverId()))
		if err != nil {
			return nil, err
		}
		if saved != nil {
			pref.AcceptRealtime = saved.AcceptRealtime
			pref.AcceptReservation = saved.AcceptReservation
			updatedAt = saved.UpdatedAt.Unix()
		}
	}
	return listenPreferenceResponse(in.GetDriverId(), pref, updatedAt), nil
}

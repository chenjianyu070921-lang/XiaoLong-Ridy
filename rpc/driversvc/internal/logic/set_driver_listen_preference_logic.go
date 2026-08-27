package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDriverListenPreferenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDriverListenPreferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDriverListenPreferenceLogic {
	return &SetDriverListenPreferenceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDriverListenPreferenceLogic) SetDriverListenPreference(in *proto.SetDriverListenPreferenceRequest) (*proto.DriverListenPreferenceResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	pref := driverListenPreference{
		AcceptRealtime:    in.GetAcceptRealtime(),
		AcceptReservation: in.GetAcceptReservation(),
	}
	if !validDriverListenPreference(pref) {
		return nil, errInvalidListenPreference
	}
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}
	if err := saveDriverListenPreference(l.ctx, l.svcCtx, in.GetDriverId(), pref); err != nil {
		return nil, err
	}
	updatedAt := time.Now().Unix()
	if l.svcCtx.OnlineStore != nil {
		state, err := l.svcCtx.OnlineStore.Get(l.ctx, in.GetDriverId())
		if err != nil {
			return nil, err
		}
		if state != nil && state.OnlineStatus == int32(DriverOnline) {
			if err := syncDispatchDriverOnlineWithPreference(l.ctx, l.svcCtx, in.GetDriverId(), state.Longitude, state.Latitude, pref); err != nil {
				return nil, err
			}
		}
	}
	return listenPreferenceResponse(in.GetDriverId(), pref, updatedAt), nil
}

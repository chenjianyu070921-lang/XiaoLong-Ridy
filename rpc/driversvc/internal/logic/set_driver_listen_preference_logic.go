package logic

import (
	"context"
	"errors"
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
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil || l.svcCtx.DriverListenPreferenceRepository == nil {
		return nil, errors.New("driver dependencies not ready")
	}
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}
	if err := saveDriverListenPreference(l.ctx, l.svcCtx, in.GetDriverId(), pref); err != nil {
		return nil, err
	}
	updatedAt := time.Now().Unix()
	// 偏好已保存到 DB，后续派单侧同步失败仅告警，不阻断主流程（避免司机端收到"设置失败"但实际已保存）
	if l.svcCtx.OnlineStore != nil {
		state, err := l.svcCtx.OnlineStore.Get(l.ctx, in.GetDriverId())
		if err != nil {
			l.Errorf("get online state for listen preference sync failed: %v", err)
			return listenPreferenceResponse(in.GetDriverId(), pref, updatedAt), nil
		}
		if state != nil && state.OnlineStatus == int32(DriverOnline) {
			if err := syncDispatchDriverOnlineWithPreference(l.ctx, l.svcCtx, in.GetDriverId(), state.Longitude, state.Latitude, pref); err != nil {
				l.Errorf("sync dispatch listen preference failed: %v", err)
			}
		}
	}
	return listenPreferenceResponse(in.GetDriverId(), pref, updatedAt), nil
}

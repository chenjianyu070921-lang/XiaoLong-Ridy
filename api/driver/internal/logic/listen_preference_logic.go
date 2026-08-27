package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type ListenPreferenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListenPreferenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListenPreferenceLogic {
	return &ListenPreferenceLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ListenPreferenceLogic) SetListenPreference(driverID int64, req *types.DriverListenPreferenceRequest) (*types.DriverListenPreferenceResponse, error) {
	if driverID <= 0 || req == nil || (!req.AcceptRealtime && !req.AcceptReservation) {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SetDriverListenPreference(l.ctx, &driversproto.SetDriverListenPreferenceRequest{
		DriverId:          driverID,
		AcceptRealtime:    req.AcceptRealtime,
		AcceptReservation: req.AcceptReservation,
	})
	if err != nil {
		return nil, err
	}
	return toDriverListenPreferenceResponse(driverID, resp), nil
}

func (l *ListenPreferenceLogic) GetListenPreference(driverID int64) (*types.DriverListenPreferenceResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetDriverListenPreference(l.ctx, &driversproto.GetDriverListenPreferenceRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	return toDriverListenPreferenceResponse(driverID, resp), nil
}

func (l *ListenPreferenceLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

func toDriverListenPreferenceResponse(driverID int64, resp *driversproto.DriverListenPreferenceResponse) *types.DriverListenPreferenceResponse {
	pref := resp.GetPreference()
	if pref == nil {
		return &types.DriverListenPreferenceResponse{DriverID: driverID, AcceptRealtime: true, AcceptReservation: true}
	}
	return &types.DriverListenPreferenceResponse{
		DriverID:          pref.GetDriverId(),
		AcceptRealtime:    pref.GetAcceptRealtime(),
		AcceptReservation: pref.GetAcceptReservation(),
		UpdatedAt:         pref.GetUpdatedAt(),
	}
}

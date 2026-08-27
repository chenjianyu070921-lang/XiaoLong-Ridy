package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

var errInvalidListenPreference = errors.New("driver listen preference must accept at least one order type")

func resolveDriverListenPreference(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, acceptRealtime, acceptReservation *bool) (driverListenPreference, error) {
	pref := defaultDriverListenPreference()
	if svcCtx != nil && svcCtx.DriverListenPreferenceRepository != nil {
		saved, err := svcCtx.DriverListenPreferenceRepository.GetByDriverID(ctx, uint64(driverID))
		if err != nil {
			return pref, err
		}
		if saved != nil {
			pref.AcceptRealtime = saved.AcceptRealtime
			pref.AcceptReservation = saved.AcceptReservation
		}
	}
	if acceptRealtime != nil {
		pref.AcceptRealtime = *acceptRealtime
	}
	if acceptReservation != nil {
		pref.AcceptReservation = *acceptReservation
	}
	if !validDriverListenPreference(pref) {
		return pref, errInvalidListenPreference
	}
	return pref, nil
}

func validDriverListenPreference(pref driverListenPreference) bool {
	return pref.AcceptRealtime || pref.AcceptReservation
}

func saveDriverListenPreference(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, pref driverListenPreference) error {
	if svcCtx == nil || svcCtx.DriverListenPreferenceRepository == nil {
		return nil
	}
	now := time.Now()
	return svcCtx.DriverListenPreferenceRepository.Upsert(ctx, &model.DriverListenPreference{
		DriverID:          uint64(driverID),
		AcceptRealtime:    pref.AcceptRealtime,
		AcceptReservation: pref.AcceptReservation,
		UpdatedAt:         now,
	})
}

func listenPreferenceResponse(driverID int64, pref driverListenPreference, updatedAt int64) *proto.DriverListenPreferenceResponse {
	return &proto.DriverListenPreferenceResponse{Preference: &proto.DriverListenPreference{
		DriverId:          driverID,
		AcceptRealtime:    pref.AcceptRealtime,
		AcceptReservation: pref.AcceptReservation,
		UpdatedAt:         updatedAt,
	}}
}

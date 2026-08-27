package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestSetListenPreferenceForwardsCurrentDriver(t *testing.T) {
	client := &fakeDriverClient{}
	logic := NewListenPreferenceLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	resp, err := logic.SetListenPreference(25, &types.DriverListenPreferenceRequest{
		AcceptRealtime:    false,
		AcceptReservation: true,
	})
	if err != nil {
		t.Fatalf("SetListenPreference() error = %v", err)
	}
	if client.setListenPreferenceRequest.GetDriverId() != 25 || client.setListenPreferenceRequest.GetAcceptRealtime() || !client.setListenPreferenceRequest.GetAcceptReservation() {
		t.Fatalf("SetListenPreference() request = %+v", client.setListenPreferenceRequest)
	}
	if resp.DriverID != 25 || resp.AcceptRealtime || !resp.AcceptReservation {
		t.Fatalf("SetListenPreference() response = %+v", resp)
	}
}

func TestGetListenPreferenceReturnsDefaultFromDriversvc(t *testing.T) {
	client := &fakeDriverClient{getListenPreferenceResponse: &driversproto.DriverListenPreferenceResponse{Preference: &driversproto.DriverListenPreference{
		DriverId:          25,
		AcceptRealtime:    true,
		AcceptReservation: true,
		UpdatedAt:         123,
	}}}
	logic := NewListenPreferenceLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	resp, err := logic.GetListenPreference(25)
	if err != nil {
		t.Fatalf("GetListenPreference() error = %v", err)
	}
	if client.getListenPreferenceRequest.GetDriverId() != 25 {
		t.Fatalf("GetListenPreference() request = %+v", client.getListenPreferenceRequest)
	}
	if resp.DriverID != 25 || !resp.AcceptRealtime || !resp.AcceptReservation || resp.UpdatedAt != 123 {
		t.Fatalf("GetListenPreference() response = %+v", resp)
	}
}

func TestSetListenPreferenceRejectsNone(t *testing.T) {
	logic := NewListenPreferenceLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})
	_, err := logic.SetListenPreference(25, &types.DriverListenPreferenceRequest{})
	if err != ErrInvalidParam {
		t.Fatalf("SetListenPreference() error = %v, want %v", err, ErrInvalidParam)
	}
}
package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func TestSetOnlineForwardsDeviceAndLocation(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewOnlineLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.SetOnline(25, &types.SetOnlineRequest{
		DeviceID:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("SetOnline() error = %v", err)
	}
	if resp.DriverID != 25 || resp.OnlineStatus != 1 {
		t.Fatalf("SetOnline() response = %+v", resp)
	}
	req := driverClient.setOnlineRequest
	if req.GetDriverId() != 25 || req.GetDeviceId() != "device-1" ||
		req.GetLongitude() != 116.397 || req.GetLatitude() != 39.908 {
		t.Fatalf("SetOnline() request = %+v", req)
	}
}

func TestSetOfflineForwardsDeviceAndLocation(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewOfflineLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.SetOffline(25, &types.SetOfflineRequest{
		DeviceID:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("SetOffline() error = %v", err)
	}
	if resp.DriverID != 25 || resp.OnlineStatus != 0 {
		t.Fatalf("SetOffline() response = %+v", resp)
	}
	req := driverClient.setOfflineRequest
	if req.GetDriverId() != 25 || req.GetDeviceId() != "device-1" ||
		req.GetLongitude() != 116.397 || req.GetLatitude() != 39.908 {
		t.Fatalf("SetOffline() request = %+v", req)
	}
}

func TestSetOnlineForwardsListenPreference(t *testing.T) {
	acceptRealtime := false
	acceptReservation := true
	driverClient := &fakeDriverClient{}
	logic := NewOnlineLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	_, err := logic.SetOnline(25, &types.SetOnlineRequest{
		DeviceID:          "device-1",
		Longitude:         116.397,
		Latitude:          39.908,
		AcceptRealtime:    &acceptRealtime,
		AcceptReservation: &acceptReservation,
	})
	if err != nil {
		t.Fatalf("SetOnline() error = %v", err)
	}
	if driverClient.setOnlineRequest.GetAcceptRealtime() || !driverClient.setOnlineRequest.GetAcceptReservation() {
		t.Fatalf("SetOnline() preference = realtime:%v reservation:%v, want false/true",
			driverClient.setOnlineRequest.GetAcceptRealtime(), driverClient.setOnlineRequest.GetAcceptReservation())
	}
}
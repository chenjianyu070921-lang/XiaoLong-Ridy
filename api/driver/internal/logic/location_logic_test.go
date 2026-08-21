package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func TestReportLocationForwardsCurrentDriverAndLocation(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.ReportLocation(25, &types.ReportLocationRequest{
		DeviceID:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("ReportLocation() error = %v", err)
	}
	if resp.DriverID != 25 || resp.OnlineStatus != 1 || resp.ReportTime != 123 {
		t.Fatalf("ReportLocation() response = %+v", resp)
	}
	req := driverClient.reportLocationRequest
	if req.GetDriverId() != 25 || req.GetDeviceId() != "device-1" ||
		req.GetLongitude() != 116.397 || req.GetLatitude() != 39.908 {
		t.Fatalf("ReportLocation() request = %+v", req)
	}
}

func TestReportLocationRejectsInvalidInput(t *testing.T) {
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})

	cases := []struct {
		name     string
		driverID int64
		req      *types.ReportLocationRequest
	}{
		{name: "missing driver", driverID: 0, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 1, Latitude: 1}},
		{name: "missing device", driverID: 25, req: &types.ReportLocationRequest{Longitude: 1, Latitude: 1}},
		{name: "bad longitude", driverID: 25, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 181, Latitude: 1}},
		{name: "bad latitude", driverID: 25, req: &types.ReportLocationRequest{DeviceID: "d", Longitude: 1, Latitude: -91}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := logic.ReportLocation(tc.driverID, tc.req); err != ErrInvalidParam {
				t.Fatalf("ReportLocation() error = %v, want %v", err, ErrInvalidParam)
			}
		})
	}
}

package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
)

type fakeLocationClient struct {
	reportLocationRequest *locationproto.ReportLocationReq
}

func (f *fakeLocationClient) ReportLocation(_ context.Context, req *locationproto.ReportLocationReq) (*locationproto.ReportLocationResp, error) {
	f.reportLocationRequest = req
	return &locationproto.ReportLocationResp{Success: true}, nil
}

func TestReportLocationAlsoWritesLocationServiceWhenConfigured(t *testing.T) {
	driverClient := &fakeDriverClient{}
	locationClient := &fakeLocationClient{}
	trajectoryRepo := &fakeTrajectoryRepository{}
	logic := NewLocationLogic(context.Background(), &svc.ServiceContext{
		DriverClient:         driverClient,
		LocationClient:       locationClient,
		TrajectoryRepository: trajectoryRepo,
	})

	_, err := logic.ReportLocation(25, &types.ReportLocationRequest{
		DeviceID:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
		OrderID:   1001,
	})
	if err != nil {
		t.Fatalf("ReportLocation() error = %v", err)
	}

	if driverClient.reportLocationRequest == nil {
		t.Fatal("driversvc ReportLocation was not called")
	}
	req := locationClient.reportLocationRequest
	if req == nil {
		t.Fatal("locationsvc ReportLocation was not called")
	}
	if req.GetDriverId() != 25 || req.GetLng() != 116.397 || req.GetLat() != 39.908 || req.GetOnlineStatus() != 1 || req.GetOrderId() != 1001 {
		t.Fatalf("locationsvc ReportLocation request = %+v", req)
	}
	if trajectoryRepo.recorded == nil {
		t.Fatal("trajectory repository was not called")
	}
	if trajectoryRepo.recorded.DriverID != 25 || trajectoryRepo.recorded.OrderID != 1001 || trajectoryRepo.recorded.Longitude != 116.397 || trajectoryRepo.recorded.Latitude != 39.908 {
		t.Fatalf("trajectory record = %+v", trajectoryRepo.recorded)
	}
}

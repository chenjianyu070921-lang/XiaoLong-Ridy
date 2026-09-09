package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	locationproto "XiaoLong-Ridy/rpc/locationsvc/locationsvc"
)

type fakeLocationClient struct {
	calls int
}

type fakeTrajectoryRepository struct {
	recorded *svc.TrajectoryRecord
}

func (f *fakeTrajectoryRepository) RecordPoint(_ context.Context, record *svc.TrajectoryRecord) error {
	copied := *record
	f.recorded = &copied
	return nil
}

func (f *fakeTrajectoryRepository) ListByOrder(_ context.Context, _, _ int64) ([]svc.TrajectoryRecord, error) {
	return nil, nil
}

func (f *fakeLocationClient) ReportLocation(_ context.Context, req *locationproto.ReportLocationReq) (*locationproto.ReportLocationResp, error) {
	f.calls++
	return &locationproto.ReportLocationResp{Success: true}, nil
}

func TestReportLocationDoesNotWriteLocationService(t *testing.T) {
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
	if locationClient.calls != 0 {
		t.Fatalf("locationsvc ReportLocation was called %d times", locationClient.calls)
	}
	if trajectoryRepo.recorded == nil {
		t.Fatal("trajectory repository was not called")
	}
	if trajectoryRepo.recorded.DriverID != 25 || trajectoryRepo.recorded.OrderID != 1001 || trajectoryRepo.recorded.Longitude != 116.397 || trajectoryRepo.recorded.Latitude != 39.908 {
		t.Fatalf("trajectory record = %+v", trajectoryRepo.recorded)
	}
}

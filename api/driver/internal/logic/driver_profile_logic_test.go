package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestGetDriverIncludesVehicleID(t *testing.T) {
	driverClient := &fakeDriverClient{
		getDriverResponse: &driversproto.GetDriverResponse{
			Driver: &driversproto.Driver{
				Id:           25,
				Phone:        "13800000001",
				RealName:     "Driver",
				Status:       driversproto.DriverStatus_DRIVER_STATUS_PENDING,
				OnlineStatus: 0,
				VehicleId:    77,
			},
		},
	}
	logic := NewDriverLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.GetDriver(25)
	if err != nil {
		t.Fatalf("GetDriver() error = %v", err)
	}
	if resp.Driver.ID != 25 || resp.Driver.VehicleID != 77 {
		t.Fatalf("GetDriver() response = %+v", resp)
	}
}

package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestCreateVehicleUsesCurrentDriverID(t *testing.T) {
	driverClient := &fakeDriverClient{}
	logic := NewVehicleLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	resp, err := logic.CreateVehicle(25, &types.CreateVehicleRequest{
		PlateNo:           " 粤b12345 ",
		Brand:             " BYD ",
		Model:             " Han ",
		Color:             "black",
		VehicleType:       1,
		RegistrationDate:  1700000000,
		InsuranceNo:       " ins-1 ",
		InsuranceExpireAt: 1800000000,
	})
	if err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	if resp.ID != 77 || resp.Status != "VEHICLE_STATUS_PENDING" || resp.CreatedAt != 123 {
		t.Fatalf("CreateVehicle() response = %+v", resp)
	}
	req := driverClient.createVehicleRequest
	if req == nil || req.GetDriverId() != 25 || req.GetPlateNo() != "粤B12345" ||
		req.GetBrand() != "BYD" || req.GetModel() != "Han" || req.GetInsuranceNo() != "ins-1" {
		t.Fatalf("CreateVehicle() request = %+v", req)
	}
	if req.GetRegistrationDate() != 1700000000 || req.GetInsuranceExpireAt() != 1800000000 {
		t.Fatalf("CreateVehicle() optional timestamps = %d/%d", req.GetRegistrationDate(), req.GetInsuranceExpireAt())
	}
}

func TestCreateVehicleRejectsInvalidInput(t *testing.T) {
	logic := NewVehicleLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})

	cases := []struct {
		name     string
		driverID int64
		req      *types.CreateVehicleRequest
	}{
		{name: "missing driver", driverID: 0, req: &types.CreateVehicleRequest{PlateNo: "A", Brand: "B", Model: "C", VehicleType: 1}},
		{name: "nil request", driverID: 25, req: nil},
		{name: "missing plate", driverID: 25, req: &types.CreateVehicleRequest{Brand: "B", Model: "C", VehicleType: 1}},
		{name: "missing brand", driverID: 25, req: &types.CreateVehicleRequest{PlateNo: "A", Model: "C", VehicleType: 1}},
		{name: "missing model", driverID: 25, req: &types.CreateVehicleRequest{PlateNo: "A", Brand: "B", VehicleType: 1}},
		{name: "bad vehicle type", driverID: 25, req: &types.CreateVehicleRequest{PlateNo: "A", Brand: "B", Model: "C"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := logic.CreateVehicle(tc.driverID, tc.req); err != ErrInvalidParam {
				t.Fatalf("CreateVehicle() error = %v, want %v", err, ErrInvalidParam)
			}
		})
	}
}

func TestGetVehicleRejectsVehicleFromAnotherDriver(t *testing.T) {
	driverClient := &fakeDriverClient{
		getVehicleResponse: &driversproto.GetVehicleResponse{
			Vehicle: &driversproto.Vehicle{Id: 77, DriverId: 99},
		},
	}
	logic := NewVehicleLogic(context.Background(), &svc.ServiceContext{DriverClient: driverClient})

	if _, err := logic.GetVehicle(25, 77); err != ErrForbiddenDriverResource {
		t.Fatalf("GetVehicle() error = %v, want %v", err, ErrForbiddenDriverResource)
	}
	req := driverClient.getVehicleRequest
	if req == nil || req.GetId() != 77 {
		t.Fatalf("GetVehicle() request = %+v", req)
	}
}

func TestGetVehicleReturnsCurrentDriverVehicle(t *testing.T) {
	logic := NewVehicleLogic(context.Background(), &svc.ServiceContext{DriverClient: &fakeDriverClient{}})

	resp, err := logic.GetVehicle(25, 77)
	if err != nil {
		t.Fatalf("GetVehicle() error = %v", err)
	}
	if resp.Vehicle.ID != 77 || resp.Vehicle.DriverID != 25 || resp.Vehicle.PlateNo != "粤B12345" {
		t.Fatalf("GetVehicle() response = %+v", resp)
	}
}

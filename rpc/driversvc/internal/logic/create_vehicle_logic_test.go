package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestCreateVehicleCreatesPendingVehicle(t *testing.T) {
	repo := &fakeVehicleRepository{}
	logic := NewCreateVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	resp, err := logic.CreateVehicle(&proto.CreateVehicleRequest{
		DriverId:          25,
		PlateNo:           " 粤b12345 ",
		Brand:             " BYD ",
		Model:             " Han ",
		Color:             " black ",
		VehicleType:       1,
		RegistrationDate:  ptrInt64(1700000000),
		InsuranceNo:       " INS-1 ",
		InsuranceExpireAt: ptrInt64(1800000000),
	})
	if err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	if resp.GetId() != 77 || resp.GetStatus() != proto.VehicleStatus_VEHICLE_STATUS_PENDING {
		t.Fatalf("CreateVehicle() response = %+v", resp)
	}
	vehicle := repo.created
	if vehicle == nil {
		t.Fatal("vehicle was not created")
	}
	if vehicle.DriverId != 25 || vehicle.PlateNo != "粤B12345" || vehicle.Brand != "BYD" ||
		vehicle.Model != "Han" || vehicle.Color != "black" || vehicle.InsuranceNo != "INS-1" ||
		vehicle.VehicleType != 1 || vehicle.Status != int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING) {
		t.Fatalf("created vehicle = %+v", vehicle)
	}
	if vehicle.RegistrationDate == nil || vehicle.RegistrationDate.Unix() != 1700000000 ||
		vehicle.InsuranceExpireAt == nil || vehicle.InsuranceExpireAt.Unix() != 1800000000 {
		t.Fatalf("created vehicle dates = %+v/%+v", vehicle.RegistrationDate, vehicle.InsuranceExpireAt)
	}
}

func TestCreateVehicleRejectsInvalidInputBeforeCreate(t *testing.T) {
	repo := &fakeVehicleRepository{}
	logic := NewCreateVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	if _, err := logic.CreateVehicle(&proto.CreateVehicleRequest{
		DriverId:    25,
		PlateNo:     "ABC123",
		Brand:       "BYD",
		Model:       "Han",
		VehicleType: 1,
	}); err == nil {
		t.Fatal("CreateVehicle() accepted invalid plate")
	}
	if repo.created != nil {
		t.Fatalf("invalid vehicle should not be created: %+v", repo.created)
	}
	if _, err := logic.CreateVehicle(nil); err == nil {
		t.Fatal("CreateVehicle() accepted nil request")
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}

type fakeVehicleRepository struct {
	created *model.DriverVehicle
}

func (f *fakeVehicleRepository) Create(_ context.Context, vehicle *model.DriverVehicle) error {
	vehicle.Id = 77
	vehicle.CreatedAt = time.Unix(123, 0)
	f.created = vehicle
	return nil
}

func (f *fakeVehicleRepository) GetByID(context.Context, uint64) (*model.DriverVehicle, error) {
	return nil, nil
}

func (f *fakeVehicleRepository) Update(context.Context, uint64, map[string]interface{}) error {
	return nil
}

func (f *fakeVehicleRepository) Delete(context.Context, *model.DriverVehicle) error {
	return nil
}

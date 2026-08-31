package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateVehicleUsesDriverIDAndUpdatesFields(t *testing.T) {
	repo := &updateVehicleRepository{
		vehicle: &model.DriverVehicle{
			Id:        77,
			DriverId:  25,
			PlateNo:   "粤B12345",
			Brand:     "BYD",
			Model:     "Han",
			Status:    int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING),
			UpdatedAt: time.Unix(123, 0),
		},
	}
	logic := NewUpdateVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	plateNo := "粤B54321"
	resp, err := logic.UpdateVehicle(&proto.UpdateVehicleRequest{
		Id:       77,
		DriverId: int64Ptr(25),
		PlateNo:  &plateNo,
	})
	if err != nil {
		t.Fatalf("UpdateVehicle() error = %v", err)
	}
	if resp.GetId() != 77 || resp.GetStatus() != proto.VehicleStatus_VEHICLE_STATUS_PENDING || resp.GetUpdatedAt() != 456 {
		t.Fatalf("UpdateVehicle() response = %+v", resp)
	}
	if _, ok := repo.updates["driver_id"]; ok {
		t.Fatalf("UpdateVehicle() should not update driver_id: %+v", repo.updates)
	}
	if repo.vehicle.DriverId != 25 || repo.vehicle.PlateNo != "粤B54321" {
		t.Fatalf("UpdateVehicle() vehicle = %+v", repo.vehicle)
	}
}

func TestUpdateVehicleRejectsOwnershipMismatch(t *testing.T) {
	repo := &updateVehicleRepository{
		vehicle: &model.DriverVehicle{
			Id:        77,
			DriverId:  25,
			Status:    int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING),
			UpdatedAt: time.Unix(123, 0),
		},
	}
	logic := NewUpdateVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	plateNo := "粤B54321"
	if _, err := logic.UpdateVehicle(&proto.UpdateVehicleRequest{
		Id:       77,
		DriverId: int64Ptr(99),
		PlateNo:  &plateNo,
	}); err == nil {
		t.Fatal("UpdateVehicle() accepted vehicle ownership mismatch")
	} else if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("UpdateVehicle() error = %v, want PermissionDenied", err)
	}
	if repo.updates != nil {
		t.Fatalf("UpdateVehicle() should not write updates for mismatched ownership: %+v", repo.updates)
	}
}

func TestUpdateVehicleRejectsDriverIDOnly(t *testing.T) {
	repo := &updateVehicleRepository{
		vehicle: &model.DriverVehicle{
			Id:        77,
			DriverId:  25,
			Status:    int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING),
			UpdatedAt: time.Unix(123, 0),
		},
	}
	logic := NewUpdateVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	if _, err := logic.UpdateVehicle(&proto.UpdateVehicleRequest{Id: 77, DriverId: int64Ptr(25)}); err == nil || err.Error() != "no updatable fields" {
		t.Fatalf("UpdateVehicle() error = %v, want %q", err, "no updatable fields")
	}
	if repo.updates != nil {
		t.Fatalf("UpdateVehicle() should not write updates for driver-id-only requests: %+v", repo.updates)
	}
}

type updateVehicleRepository struct {
	vehicle *model.DriverVehicle
	updates map[string]interface{}
}

func (r *updateVehicleRepository) Create(context.Context, *model.DriverVehicle) error { return nil }

func (r *updateVehicleRepository) GetByID(_ context.Context, id uint64) (*model.DriverVehicle, error) {
	if r.vehicle == nil || r.vehicle.Id != id {
		return nil, errors.New("vehicle not found")
	}
	return r.vehicle, nil
}

func (r *updateVehicleRepository) GetByDriverID(context.Context, uint64) (*model.DriverVehicle, error) {
	return nil, errors.New("not implemented")
}

func (r *updateVehicleRepository) Update(_ context.Context, id uint64, updates map[string]interface{}) error {
	if r.vehicle == nil || r.vehicle.Id != id {
		return errors.New("vehicle not found")
	}
	r.updates = updates
	if v, ok := updates["plate_no"].(string); ok {
		r.vehicle.PlateNo = v
	}
	if v, ok := updates["brand"].(string); ok {
		r.vehicle.Brand = v
	}
	if v, ok := updates["model"].(string); ok {
		r.vehicle.Model = v
	}
	if v, ok := updates["status"].(int8); ok {
		r.vehicle.Status = v
	}
	r.vehicle.UpdatedAt = time.Unix(456, 0)
	return nil
}

func (r *updateVehicleRepository) Delete(context.Context, *model.DriverVehicle) error { return nil }

func int64Ptr(v int64) *int64 { return &v }

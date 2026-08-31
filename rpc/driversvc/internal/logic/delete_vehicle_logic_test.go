package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteVehicleRequiresDriverID(t *testing.T) {
	logic := NewDeleteVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
		Id:       77,
		DriverId: 25,
	}}})

	if _, err := logic.DeleteVehicle(&proto.DeleteVehicleRequest{Id: 77}); err == nil {
		t.Fatal("DeleteVehicle() accepted missing driver id")
	}
}

func TestDeleteVehicleRejectsOwnershipMismatch(t *testing.T) {
	repo := &updateVehicleRepository{vehicle: &model.DriverVehicle{
		Id:        77,
		DriverId:  25,
		UpdatedAt: time.Unix(123, 0),
	}}
	logic := NewDeleteVehicleLogic(context.Background(), &svc.ServiceContext{DriverVehicleRepository: repo})

	if _, err := logic.DeleteVehicle(&proto.DeleteVehicleRequest{Id: 77, DriverId: 26}); err == nil {
		t.Fatal("DeleteVehicle() accepted vehicle ownership mismatch")
	} else if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("DeleteVehicle() error = %v, want PermissionDenied", err)
	}
}

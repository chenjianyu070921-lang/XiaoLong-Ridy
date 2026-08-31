package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUploadCertificationRequiresVehicleID(t *testing.T) {
	logic := NewUploadCertificationLogic(context.Background(), &svc.ServiceContext{})

	if _, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		IdCardFront: "not-decoded-before-vehicle-check",
	}); err == nil {
		t.Fatal("UploadCertification() accepted missing vehicle id")
	}
}

func TestUploadCertificationRejectsVehicleOwnershipMismatch(t *testing.T) {
	logic := NewUploadCertificationLogic(context.Background(), &svc.ServiceContext{
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 26,
		}},
	})

	if _, err := logic.UploadCertification(&proto.UploadCertificationRequest{
		DriverId:    25,
		VehicleId:   77,
		IdCardFront: "not-decoded-before-ownership-check",
	}); err == nil {
		t.Fatal("UploadCertification() accepted vehicle ownership mismatch")
	} else if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("UploadCertification() error = %v, want PermissionDenied", err)
	}
}

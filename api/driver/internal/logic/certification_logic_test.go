package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestUploadCertificationRejectsVehicleFromAnotherDriver(t *testing.T) {
	client := &fakeDriverClient{
		getVehicleResponse: &driversproto.GetVehicleResponse{Vehicle: &driversproto.Vehicle{
			Id:       77,
			DriverId: 26,
		}},
	}
	logic := NewCertificationLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	_, err := logic.UploadCertification(25, &types.UploadCertificationRequest{
		VehicleID:       77,
		IdCardNo:        "11010119900101153X",
		RealName:        "张三",
		DriverLicenseNo: "DL10000001",
	})

	if err != ErrForbiddenDriverResource {
		t.Fatalf("UploadCertification() error = %v, want %v", err, ErrForbiddenDriverResource)
	}
	if client.uploadCertificationRequest != nil {
		t.Fatalf("UploadCertification() should not forward mismatched vehicle: %+v", client.uploadCertificationRequest)
	}
}

func TestUploadCertificationRequiresVehicleID(t *testing.T) {
	client := &fakeDriverClient{}
	logic := NewCertificationLogic(context.Background(), &svc.ServiceContext{DriverClient: client})

	if _, err := logic.UploadCertification(25, &types.UploadCertificationRequest{
		IdCardNo:        "11010119900101153X",
		RealName:        "张三",
		DriverLicenseNo: "DL10000001",
	}); err != ErrInvalidParam {
		t.Fatalf("UploadCertification() error = %v, want %v", err, ErrInvalidParam)
	}
	if client.getVehicleRequest != nil || client.uploadCertificationRequest != nil {
		t.Fatalf("UploadCertification() should not call downstream without vehicle id: get=%+v upload=%+v", client.getVehicleRequest, client.uploadCertificationRequest)
	}
}

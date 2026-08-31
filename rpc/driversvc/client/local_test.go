package client

import (
	"context"
	"testing"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestLocalClientGetDriverIncludesVehicleAndCertification(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()

	created, err := cli.CreateDriver(ctx, &driversproto.CreateDriverRequest{
		Phone:           "13800138000",
		RealName:        "driver",
		IdCardNo:        "440300199001011234",
		DriverLicenseNo: "DL-25",
	})
	if err != nil {
		t.Fatalf("CreateDriver() error = %v", err)
	}
	vehicle, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    created.GetId(),
		PlateNo:     "粤B12345",
		Brand:       "BYD",
		Model:       "Han",
		Color:       "black",
		VehicleType: 1,
		InsuranceNo: "INS-1",
	})
	if err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	cert, err := cli.UploadCertification(ctx, &driversproto.UploadCertificationRequest{
		DriverId:       created.GetId(),
		VehicleId:      vehicle.GetId(),
		IdCardFront:    "front",
		IdCardBack:     "back",
		DriverLicense:  "license",
		VehicleLicense: "vehicle",
	})
	if err != nil {
		t.Fatalf("UploadCertification() error = %v", err)
	}

	detail, err := cli.GetDriver(ctx, &driversproto.GetDriverRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetDriver() error = %v", err)
	}
	driver := detail.GetDriver()
	if driver.GetVehicleId() != vehicle.GetId() || driver.GetPlateNo() != "粤B12345" ||
		driver.GetVehicleStatus() != int32(driversproto.VehicleStatus_VEHICLE_STATUS_PENDING) ||
		driver.GetCertificationId() != cert.GetId() || driver.GetAuditStatus() != 1 {
		t.Fatalf("GetDriver() aggregate fields = %+v", driver)
	}

	list, err := cli.ListDrivers(ctx, &driversproto.ListDriversRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if len(list.GetDrivers()) != 1 {
		t.Fatalf("ListDrivers() length = %d", len(list.GetDrivers()))
	}
	driver = list.GetDrivers()[0]
	if driver.GetVehicleId() != vehicle.GetId() || driver.GetPlateNo() != "粤B12345" ||
		driver.GetCertificationId() != cert.GetId() || driver.GetAuditStatus() != 1 {
		t.Fatalf("ListDrivers() aggregate fields = %+v", driver)
	}
}

func TestLocalClientListDriversKeywordMatchesOlderVehiclePlate(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()

	created, err := cli.CreateDriver(ctx, &driversproto.CreateDriverRequest{
		Phone:           "13800138001",
		RealName:        "driver",
		IdCardNo:        "440300199001011235",
		DriverLicenseNo: "DL-26",
	})
	if err != nil {
		t.Fatalf("CreateDriver() error = %v", err)
	}
	if _, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    created.GetId(),
		PlateNo:     "OLD-PLATE",
		Brand:       "BYD",
		Model:       "Han",
		Color:       "black",
		VehicleType: 1,
		InsuranceNo: "INS-OLD",
	}); err != nil {
		t.Fatalf("CreateVehicle(old) error = %v", err)
	}
	latest, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    created.GetId(),
		PlateNo:     "NEW-PLATE",
		Brand:       "BYD",
		Model:       "Tang",
		Color:       "white",
		VehicleType: 1,
		InsuranceNo: "INS-NEW",
	})
	if err != nil {
		t.Fatalf("CreateVehicle(new) error = %v", err)
	}

	list, err := cli.ListDrivers(ctx, &driversproto.ListDriversRequest{Page: 1, PageSize: 20, Keyword: "OLD-PLATE"})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if len(list.GetDrivers()) != 1 || list.GetDrivers()[0].GetId() != created.GetId() {
		t.Fatalf("ListDrivers() drivers = %+v", list.GetDrivers())
	}
	if list.GetDrivers()[0].GetVehicleId() != latest.GetId() || list.GetDrivers()[0].GetPlateNo() != "NEW-PLATE" {
		t.Fatalf("ListDrivers() aggregate should still use latest vehicle, got %+v", list.GetDrivers()[0])
	}
}

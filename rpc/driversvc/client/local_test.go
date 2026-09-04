package client

import (
	"context"
	"testing"

	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLocalClientGetDriverIncludesVehicleAndCertification(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()

	created, err := cli.CreateDriver(ctx, &driversproto.CreateDriverRequest{
		Phone:           "13800138000",
		PasswordHash:    "Driver@123",
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
		PasswordHash:    "Driver@123",
		RealName:        "driver",
		IdCardNo:        "440300199001011235",
		DriverLicenseNo: "DL-26",
	})
	if err != nil {
		t.Fatalf("CreateDriver() error = %v", err)
	}
	if _, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    created.GetId(),
		PlateNo:     "粤B11111",
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
		PlateNo:     "粤B22222",
		Brand:       "BYD",
		Model:       "Tang",
		Color:       "white",
		VehicleType: 1,
		InsuranceNo: "INS-NEW",
	})
	if err != nil {
		t.Fatalf("CreateVehicle(new) error = %v", err)
	}

	list, err := cli.ListDrivers(ctx, &driversproto.ListDriversRequest{Page: 1, PageSize: 20, Keyword: "粤B11111"})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if len(list.GetDrivers()) != 1 || list.GetDrivers()[0].GetId() != created.GetId() {
		t.Fatalf("ListDrivers() drivers = %+v", list.GetDrivers())
	}
	if list.GetDrivers()[0].GetVehicleId() != latest.GetId() || list.GetDrivers()[0].GetPlateNo() != "粤B22222" {
		t.Fatalf("ListDrivers() aggregate should still use latest vehicle, got %+v", list.GetDrivers()[0])
	}
}

func TestLocalClientUploadCertificationRequiresVehicleID(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()
	driver := createLocalDriverForTest(t, ctx, cli, "13800138002")

	_, err := cli.UploadCertification(ctx, &driversproto.UploadCertificationRequest{
		DriverId:      driver.GetId(),
		IdCardFront:   "front",
		IdCardBack:    "back",
		DriverLicense: "license",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("UploadCertification() code = %v, want %v (err=%v)", status.Code(err), codes.InvalidArgument, err)
	}
}

func TestLocalClientRegisterDriverRejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		req  *driversproto.CreateDriverRequest
	}{
		{name: "invalid phone", req: &driversproto.CreateDriverRequest{Phone: "123", PasswordHash: "Driver@123", RealName: "driver", IdCardNo: "440300199001011234", DriverLicenseNo: "DL-25"}},
		{name: "short password", req: &driversproto.CreateDriverRequest{Phone: "13800138006", PasswordHash: "1234567", RealName: "driver", IdCardNo: "440300199001011234", DriverLicenseNo: "DL-25"}},
		{name: "empty real name", req: &driversproto.CreateDriverRequest{Phone: "13800138006", PasswordHash: "Driver@123", RealName: "   ", IdCardNo: "440300199001011234", DriverLicenseNo: "DL-25"}},
		{name: "invalid id card", req: &driversproto.CreateDriverRequest{Phone: "13800138006", PasswordHash: "Driver@123", RealName: "driver", IdCardNo: "44030019900101123Y", DriverLicenseNo: "DL-25"}},
		{name: "empty driver license", req: &driversproto.CreateDriverRequest{Phone: "13800138006", PasswordHash: "Driver@123", RealName: "driver", IdCardNo: "440300199001011234", DriverLicenseNo: "   "}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli := NewLocalClient()
			if _, err := cli.RegisterDriver(ctx, tc.req); err == nil {
				t.Fatal("RegisterDriver() accepted invalid input")
			}
		})
	}
}

func TestLocalClientCreateVehicleValidatesAndNormalizesInput(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()
	driver := createLocalDriverForTest(t, ctx, cli, "13800138007")

	cases := []struct {
		name string
		req  *driversproto.CreateVehicleRequest
	}{
		{name: "invalid plate", req: &driversproto.CreateVehicleRequest{DriverId: driver.GetId(), PlateNo: "TEST-PLATE", Brand: "BYD", Model: "Han", VehicleType: 1}},
		{name: "invalid type", req: &driversproto.CreateVehicleRequest{DriverId: driver.GetId(), PlateNo: "粤B12345", Brand: "BYD", Model: "Han", VehicleType: 0}},
		{name: "missing brand", req: &driversproto.CreateVehicleRequest{DriverId: driver.GetId(), PlateNo: "粤B12345", Model: "Han", VehicleType: 1}},
		{name: "missing model", req: &driversproto.CreateVehicleRequest{DriverId: driver.GetId(), PlateNo: "粤B12345", Brand: "BYD", VehicleType: 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := cli.CreateVehicle(ctx, tc.req); err == nil {
				t.Fatal("CreateVehicle() accepted invalid input")
			}
		})
	}

	resp, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    driver.GetId(),
		PlateNo:     " 粤b12345 ",
		Brand:       " BYD ",
		Model:       " Han ",
		Color:       " black ",
		VehicleType: 1,
		InsuranceNo: " INS-1 ",
	})
	if err != nil {
		t.Fatalf("CreateVehicle() valid input error = %v", err)
	}
	vehicle := cli.vehicles[uint64(resp.GetId())]
	if vehicle.GetPlateNo() != "粤B12345" || vehicle.GetBrand() != "BYD" ||
		vehicle.GetModel() != "Han" || vehicle.GetColor() != "black" || vehicle.GetInsuranceNo() != "INS-1" {
		t.Fatalf("CreateVehicle() did not normalize fields: %+v", vehicle)
	}
}

func TestLocalClientUpdateVehicleRequiresOwnerAndRejectsMismatch(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()
	owner := createLocalDriverForTest(t, ctx, cli, "13800138008")
	other := createLocalDriverForTest(t, ctx, cli, "13800138009")
	vehicle := createLocalVehicleForTest(t, ctx, cli, owner.GetId())
	brand := "Denza"

	_, err := cli.UpdateVehicle(ctx, &driversproto.UpdateVehicleRequest{
		Id:    vehicle.GetId(),
		Brand: &brand,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("UpdateVehicle() without driver id code = %v, want %v (err=%v)", status.Code(err), codes.InvalidArgument, err)
	}

	otherDriverID := other.GetId()
	_, err = cli.UpdateVehicle(ctx, &driversproto.UpdateVehicleRequest{
		Id:       vehicle.GetId(),
		DriverId: &otherDriverID,
		Brand:    &brand,
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("UpdateVehicle() mismatch code = %v, want %v (err=%v)", status.Code(err), codes.PermissionDenied, err)
	}
	if got := cli.vehicles[uint64(vehicle.GetId())].GetDriverId(); got != owner.GetId() {
		t.Fatalf("UpdateVehicle() reassigned owner to %d, want %d", got, owner.GetId())
	}

	ownerID := owner.GetId()
	resp, err := cli.UpdateVehicle(ctx, &driversproto.UpdateVehicleRequest{
		Id:       vehicle.GetId(),
		DriverId: &ownerID,
		Brand:    &brand,
	})
	if err != nil {
		t.Fatalf("UpdateVehicle() owner update error = %v", err)
	}
	if resp.GetId() != vehicle.GetId() || cli.vehicles[uint64(vehicle.GetId())].GetBrand() != brand {
		t.Fatalf("UpdateVehicle() owner update failed: resp=%+v vehicle=%+v", resp, cli.vehicles[uint64(vehicle.GetId())])
	}
}

func TestLocalClientUploadCertificationRejectsVehicleFromAnotherDriver(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()
	owner := createLocalDriverForTest(t, ctx, cli, "13800138003")
	other := createLocalDriverForTest(t, ctx, cli, "13800138004")
	vehicle := createLocalVehicleForTest(t, ctx, cli, owner.GetId())

	_, err := cli.UploadCertification(ctx, &driversproto.UploadCertificationRequest{
		DriverId:       other.GetId(),
		VehicleId:      vehicle.GetId(),
		IdCardFront:    "front",
		IdCardBack:     "back",
		DriverLicense:  "license",
		VehicleLicense: "vehicle",
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("UploadCertification() code = %v, want %v (err=%v)", status.Code(err), codes.PermissionDenied, err)
	}
}

func TestLocalClientSetDriverOnlineRequiresActiveApprovedDriverAndVehicle(t *testing.T) {
	ctx := context.Background()
	cli := NewLocalClient()
	driver := createLocalDriverForTest(t, ctx, cli, "13800138005")

	_, err := cli.SetDriverOnline(ctx, &driversproto.SetDriverOnlineRequest{
		DriverId:  driver.GetId(),
		DeviceId:  "device-1",
		Longitude: 113.93,
		Latitude:  22.54,
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("SetDriverOnline() without approval code = %v, want %v (err=%v)", status.Code(err), codes.PermissionDenied, err)
	}

	vehicle := createLocalVehicleForTest(t, ctx, cli, driver.GetId())
	cert, err := cli.UploadCertification(ctx, &driversproto.UploadCertificationRequest{
		DriverId:       driver.GetId(),
		VehicleId:      vehicle.GetId(),
		IdCardFront:    "front",
		IdCardBack:     "back",
		DriverLicense:  "license",
		VehicleLicense: "vehicle",
	})
	if err != nil {
		t.Fatalf("UploadCertification() error = %v", err)
	}

	cli.mu.Lock()
	cli.drivers[uint64(driver.GetId())].Status = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	cli.certByID[uint64(cert.GetId())].AuditStatus = 2
	cli.mu.Unlock()

	_, err = cli.SetDriverOnline(ctx, &driversproto.SetDriverOnlineRequest{
		DriverId:  driver.GetId(),
		DeviceId:  "device-1",
		Longitude: 113.93,
		Latitude:  22.54,
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("SetDriverOnline() with pending vehicle code = %v, want %v (err=%v)", status.Code(err), codes.PermissionDenied, err)
	}

	cli.mu.Lock()
	cli.vehicles[uint64(vehicle.GetId())].Status = driversproto.VehicleStatus_VEHICLE_STATUS_NORMAL
	cli.mu.Unlock()

	resp, err := cli.SetDriverOnline(ctx, &driversproto.SetDriverOnlineRequest{
		DriverId:  driver.GetId(),
		DeviceId:  "device-1",
		Longitude: 113.93,
		Latitude:  22.54,
	})
	if err != nil {
		t.Fatalf("SetDriverOnline() after approval error = %v", err)
	}
	if resp.GetOnlineStatus() != 1 {
		t.Fatalf("SetDriverOnline() online status = %d, want 1", resp.GetOnlineStatus())
	}
}

func createLocalDriverForTest(t *testing.T, ctx context.Context, cli *LocalClient, phone string) *driversproto.CreateDriverResponse {
	t.Helper()
	idSuffix := phone[len(phone)-4:]
	driver, err := cli.CreateDriver(ctx, &driversproto.CreateDriverRequest{
		Phone:           phone,
		PasswordHash:    "Driver@123",
		RealName:        "driver",
		IdCardNo:        "44030019900101" + idSuffix,
		DriverLicenseNo: phone + "DL",
	})
	if err != nil {
		t.Fatalf("CreateDriver() error = %v", err)
	}
	return driver
}

func createLocalVehicleForTest(t *testing.T, ctx context.Context, cli *LocalClient, driverID int64) *driversproto.CreateVehicleResponse {
	t.Helper()
	vehicle, err := cli.CreateVehicle(ctx, &driversproto.CreateVehicleRequest{
		DriverId:    driverID,
		PlateNo:     "粤B92345",
		Brand:       "BYD",
		Model:       "Han",
		Color:       "black",
		VehicleType: 1,
		InsuranceNo: "INS-TEST",
	})
	if err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	return vehicle
}

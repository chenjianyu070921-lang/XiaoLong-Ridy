package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestGetDriverReturnsBaseProfileWhenOnlineStoreUnavailable(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mr.Close()

	logic := NewGetDriverLogic(context.Background(), &svc.ServiceContext{
		DriverRepository: &getDriverFakeDriverRepository{driver: &model.Driver{
			Id:              25,
			Phone:           "13800138000",
			RealName:        "driver",
			IdCardNo:        "440300199001011234",
			DriverLicenseNo: "DL-25",
			Status:          int8(proto.DriverStatus_DRIVER_STATUS_NORMAL),
			OnlineStatus:    1,
			CreatedAt:       time.Unix(100, 0),
			UpdatedAt:       time.Unix(200, 0),
		}},
		OnlineStore: onlinestore.NewStore(rdb, time.Minute),
	})

	resp, err := logic.GetDriver(&proto.GetDriverRequest{Id: 25})
	if err != nil {
		t.Fatalf("GetDriver() error = %v", err)
	}
	driver := resp.GetDriver()
	if driver.GetId() != 25 || driver.GetOnlineStatus() != 1 || driver.GetPhone() != "13800138000" {
		t.Fatalf("GetDriver() driver = %+v", driver)
	}
}

func TestListDriversSkipsBrokenAuxiliaryAggregates(t *testing.T) {
	logic := NewListDriversLogic(context.Background(), &svc.ServiceContext{
		DriverRepository: &getDriverFakeDriverRepository{list: []*model.Driver{
			{
				Id:              25,
				Phone:           "13800138000",
				RealName:        "driver",
				IdCardNo:        "440300199001011234",
				DriverLicenseNo: "DL-25",
				Status:          int8(proto.DriverStatus_DRIVER_STATUS_NORMAL),
				OnlineStatus:    1,
				CreatedAt:       time.Unix(100, 0),
				UpdatedAt:       time.Unix(200, 0),
			},
		}},
		DriverVehicleRepository: getDriverBrokenVehicleRepository{},
		CertificationRepository: getDriverBrokenCertificationRepository{},
	})

	resp, err := logic.ListDrivers(&proto.ListDriversRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if resp.GetTotal() != 1 || len(resp.GetDrivers()) != 1 {
		t.Fatalf("ListDrivers() response = %+v", resp)
	}
	if resp.GetDrivers()[0].GetId() != 25 {
		t.Fatalf("ListDrivers() driver = %+v", resp.GetDrivers()[0])
	}
}

// TestListDriversBuildsAdminAggregateFields 验证司机列表为管理后台聚合司机端权威字段。
// 入参为后台分页查询条件，返回值应包含司机基础资料、Redis 在线状态、车辆信息和认证信息。
func TestListDriversBuildsAdminAggregateFields(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	if err := onlinestore.NewStore(rdb, time.Minute).SetOnline(context.Background(), 25, "device-admin-test", 116.397, 39.908); err != nil {
		t.Fatalf("SetOnline() error = %v", err)
	}

	logic := NewListDriversLogic(context.Background(), &svc.ServiceContext{
		DriverRepository: &getDriverFakeDriverRepository{list: []*model.Driver{
			{
				Id:              25,
				Phone:           "13800138000",
				RealName:        "driver",
				IdCardNo:        "440300199001011234",
				DriverLicenseNo: "DL-25",
				Status:          int8(proto.DriverStatus_DRIVER_STATUS_NORMAL),
				OnlineStatus:    0,
				CreatedAt:       time.Unix(100, 0),
				UpdatedAt:       time.Unix(200, 0),
			},
		}},
		OnlineStore:             onlinestore.NewStore(rdb, time.Minute),
		DriverVehicleRepository: getDriverVehicleRepository{vehicle: &model.DriverVehicle{Id: 8001, DriverId: 25, PlateNo: "京A12345", Status: int8(proto.VehicleStatus_VEHICLE_STATUS_NORMAL)}},
		CertificationRepository: getDriverCertificationRepository{certification: &model.DriverCertification{Id: 9001, DriverId: 25, VehicleId: 8001, AuditStatus: AuditStatusPassed, AuditRemark: "审核通过"}},
	})

	resp, err := logic.ListDrivers(&proto.ListDriversRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListDrivers() error = %v", err)
	}
	if len(resp.GetDrivers()) != 1 {
		t.Fatalf("ListDrivers() got %d drivers, want 1", len(resp.GetDrivers()))
	}
	driver := resp.GetDrivers()[0]
	if driver.GetOnlineStatus() != onlinestore.Online {
		t.Fatalf("driver online status = %d, want Redis online status %d", driver.GetOnlineStatus(), onlinestore.Online)
	}
	if driver.GetVehicleId() != 8001 || driver.GetPlateNo() != "京A12345" || driver.GetVehicleStatus() != int32(proto.VehicleStatus_VEHICLE_STATUS_NORMAL) {
		t.Fatalf("driver vehicle aggregate = %+v, want vehicle fields from repository", driver)
	}
	if driver.GetCertificationId() != 9001 || driver.GetAuditStatus() != int32(AuditStatusPassed) || driver.GetAuditRemark() != "审核通过" {
		t.Fatalf("driver certification aggregate = %+v, want certification fields from repository", driver)
	}
}

type getDriverFakeDriverRepository struct {
	driver *model.Driver
	list   []*model.Driver
}

func (f *getDriverFakeDriverRepository) Create(context.Context, *model.Driver) error {
	return nil
}

func (f *getDriverFakeDriverRepository) GetByID(context.Context, uint64) (*model.Driver, error) {
	if f.driver == nil {
		return nil, repository.ErrDriverNotFound
	}
	return f.driver, nil
}

func (f *getDriverFakeDriverRepository) GetByPhone(context.Context, string) (*model.Driver, error) {
	return nil, repository.ErrDriverNotFound
}

func (f *getDriverFakeDriverRepository) List(context.Context, repository.DriverListFilter) ([]*model.Driver, int64, error) {
	return f.list, int64(len(f.list)), nil
}

func (f *getDriverFakeDriverRepository) ListNearbyDrivers(context.Context, repository.NearbyDriverFilter) ([]*model.DriverLocation, error) {
	return nil, nil
}

func (f *getDriverFakeDriverRepository) UpsertLocation(context.Context, *model.DriverLocation) error {
	return nil
}

func (f *getDriverFakeDriverRepository) UpdateLocationStatus(context.Context, uint64, int8) error {
	return nil
}

func (f *getDriverFakeDriverRepository) UpdateStatusAndLocation(context.Context, uint64, int8) error {
	return nil
}

func (f *getDriverFakeDriverRepository) GetDriverScore(context.Context, uint64) (*model.DriverScore, error) {
	return nil, nil
}

func (f *getDriverFakeDriverRepository) Update(context.Context, uint64, map[string]interface{}) error {
	return nil
}

func (f *getDriverFakeDriverRepository) Delete(context.Context, *model.Driver) error {
	return nil
}

type getDriverBrokenVehicleRepository struct{}

func (getDriverBrokenVehicleRepository) Create(context.Context, *model.DriverVehicle) error {
	return nil
}
func (getDriverBrokenVehicleRepository) GetByID(context.Context, uint64) (*model.DriverVehicle, error) {
	return nil, repository.ErrVehicleNotFound
}
func (getDriverBrokenVehicleRepository) GetByDriverID(context.Context, uint64) (*model.DriverVehicle, error) {
	return nil, errors.New("vehicle db unavailable")
}
func (getDriverBrokenVehicleRepository) Update(context.Context, uint64, map[string]interface{}) error {
	return nil
}
func (getDriverBrokenVehicleRepository) Delete(context.Context, *model.DriverVehicle) error {
	return nil
}

type getDriverVehicleRepository struct {
	vehicle *model.DriverVehicle
}

func (r getDriverVehicleRepository) Create(context.Context, *model.DriverVehicle) error {
	return nil
}
func (r getDriverVehicleRepository) GetByID(context.Context, uint64) (*model.DriverVehicle, error) {
	if r.vehicle == nil {
		return nil, repository.ErrVehicleNotFound
	}
	return r.vehicle, nil
}
func (r getDriverVehicleRepository) GetByDriverID(context.Context, uint64) (*model.DriverVehicle, error) {
	if r.vehicle == nil {
		return nil, repository.ErrVehicleNotFound
	}
	return r.vehicle, nil
}
func (r getDriverVehicleRepository) Update(context.Context, uint64, map[string]interface{}) error {
	return nil
}
func (r getDriverVehicleRepository) Delete(context.Context, *model.DriverVehicle) error {
	return nil
}

type getDriverBrokenCertificationRepository struct{}

func (getDriverBrokenCertificationRepository) Upsert(context.Context, *model.DriverCertification) (*model.DriverCertification, error) {
	return nil, nil
}
func (getDriverBrokenCertificationRepository) GetByDriverID(context.Context, uint64) (*model.DriverCertification, error) {
	return nil, errors.New("certification db unavailable")
}
func (getDriverBrokenCertificationRepository) UpdateAudit(context.Context, int64, int64, string, int8) error {
	return nil
}
func (getDriverBrokenCertificationRepository) AdminList(context.Context, repository.AdminCertificationFilter) ([]*repository.AdminCertificationRow, int64, error) {
	return nil, 0, nil
}
func (getDriverBrokenCertificationRepository) AdminGetByID(context.Context, uint64) (*repository.AdminCertificationRow, error) {
	return nil, repository.ErrCertificationNotFound
}

type getDriverCertificationRepository struct {
	certification *model.DriverCertification
}

func (r getDriverCertificationRepository) Upsert(context.Context, *model.DriverCertification) (*model.DriverCertification, error) {
	return r.certification, nil
}
func (r getDriverCertificationRepository) GetByDriverID(context.Context, uint64) (*model.DriverCertification, error) {
	if r.certification == nil {
		return nil, repository.ErrCertificationNotFound
	}
	return r.certification, nil
}
func (r getDriverCertificationRepository) UpdateAudit(context.Context, int64, int64, string, int8) error {
	return nil
}
func (getDriverCertificationRepository) AdminList(context.Context, repository.AdminCertificationFilter) ([]*repository.AdminCertificationRow, int64, error) {
	return nil, 0, nil
}
func (getDriverCertificationRepository) AdminGetByID(context.Context, uint64) (*repository.AdminCertificationRow, error) {
	return nil, repository.ErrCertificationNotFound
}

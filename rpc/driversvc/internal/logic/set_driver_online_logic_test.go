package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSetDriverOnlineClearsBusyAndMarksOnline(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	if err := rdb.SAdd(context.Background(), constants.RedisDriverBusy, "25").Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}

	repo := &setDriverOnlineRepository{
		driver: &model.Driver{Id: 25, Status: int8(proto.DriverStatus_DRIVER_STATUS_NORMAL)},
	}
	logic := NewSetDriverOnlineLogic(context.Background(), &svc.ServiceContext{
		DriverRepository:        repo,
		CertificationRepository: &setDriverOnlineCertificationRepository{certification: &model.DriverCertification{DriverId: 25, VehicleId: 77, AuditStatus: AuditStatusPassed}},
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 25,
			Status:   int8(proto.VehicleStatus_VEHICLE_STATUS_NORMAL),
		}},
		OnlineStore: onlinestore.NewStore(rdb, 0),
		RedisClient: rdb,
	})

	resp, err := logic.SetDriverOnline(&proto.SetDriverOnlineRequest{
		DriverId:  25,
		DeviceId:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("SetDriverOnline() error = %v", err)
	}
	if resp.GetDriverId() != 25 || resp.GetOnlineStatus() != int32(DriverOnline) {
		t.Fatalf("SetDriverOnline() response = %+v", resp)
	}

	busy, err := rdb.SIsMember(context.Background(), constants.RedisDriverBusy, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember(busy) error = %v", err)
	}
	if busy {
		t.Fatalf("driver should be removed from %s", constants.RedisDriverBusy)
	}
	online, err := rdb.SIsMember(context.Background(), constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember(online) error = %v", err)
	}
	if !online {
		t.Fatalf("driver should be added to %s", constants.RedisDriverOnline)
	}
	if repo.updatedStatus != DriverOnline {
		t.Fatalf("driver status update = %d, want %d", repo.updatedStatus, DriverOnline)
	}
	if repo.location == nil || repo.location.DriverID != 25 || repo.location.OnlineStatus != DriverOnline {
		t.Fatalf("location update = %+v", repo.location)
	}
}

func TestSetDriverOnlineRejectsUnapprovedCertification(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := &setDriverOnlineRepository{
		driver: &model.Driver{Id: 25, Status: int8(proto.DriverStatus_DRIVER_STATUS_NORMAL)},
	}
	certRepo := &setDriverOnlineCertificationRepository{
		certification: &model.DriverCertification{DriverId: 25, AuditStatus: AuditStatusPending},
	}
	logic := NewSetDriverOnlineLogic(context.Background(), &svc.ServiceContext{
		DriverRepository:        repo,
		CertificationRepository: certRepo,
		OnlineStore:             onlinestore.NewStore(rdb, 0),
		RedisClient:             rdb,
	})

	if _, err := logic.SetDriverOnline(&proto.SetDriverOnlineRequest{
		DriverId:  25,
		DeviceId:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	}); err == nil {
		t.Fatal("SetDriverOnline() accepted unapproved certification")
	}
	busy, err := rdb.SIsMember(context.Background(), constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember(online) error = %v", err)
	}
	if busy {
		t.Fatalf("driver should not be added to %s", constants.RedisDriverOnline)
	}
	if repo.location != nil || repo.updatedStatus != 0 {
		t.Fatalf("driver repository should not be updated: location=%+v status=%d", repo.location, repo.updatedStatus)
	}
}

func TestSetDriverOnlineRejectsDriverWithoutNormalVehicle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := &setDriverOnlineRepository{
		driver: &model.Driver{Id: 25, Status: int8(proto.DriverStatus_DRIVER_STATUS_NORMAL)},
	}
	logic := NewSetDriverOnlineLogic(context.Background(), &svc.ServiceContext{
		DriverRepository: repo,
		CertificationRepository: &setDriverOnlineCertificationRepository{
			certification: &model.DriverCertification{DriverId: 25, VehicleId: 77, AuditStatus: AuditStatusPassed},
		},
		DriverVehicleRepository: &updateVehicleRepository{vehicle: &model.DriverVehicle{
			Id:       77,
			DriverId: 25,
			Status:   int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING),
		}},
		OnlineStore: onlinestore.NewStore(rdb, 0),
		RedisClient: rdb,
	})

	if _, err := logic.SetDriverOnline(&proto.SetDriverOnlineRequest{
		DriverId:  25,
		DeviceId:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	}); err == nil {
		t.Fatal("SetDriverOnline() accepted driver without normal vehicle")
	}
	online, err := rdb.SIsMember(context.Background(), constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember(online) error = %v", err)
	}
	if online {
		t.Fatalf("driver should not be added to %s", constants.RedisDriverOnline)
	}
	if repo.location != nil || repo.updatedStatus != 0 {
		t.Fatalf("driver repository should not be updated: location=%+v status=%d", repo.location, repo.updatedStatus)
	}
}

func TestSetDriverOnlineRejectsInactiveDriver(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := &setDriverOnlineRepository{
		driver: &model.Driver{Id: 25, Status: int8(proto.DriverStatus_DRIVER_STATUS_PENDING)},
	}
	logic := NewSetDriverOnlineLogic(context.Background(), &svc.ServiceContext{
		DriverRepository:        repo,
		CertificationRepository: &setDriverOnlineCertificationRepository{certification: &model.DriverCertification{DriverId: 25, AuditStatus: AuditStatusPassed}},
		OnlineStore:             onlinestore.NewStore(rdb, 0),
		RedisClient:             rdb,
	})

	if _, err := logic.SetDriverOnline(&proto.SetDriverOnlineRequest{
		DriverId:  25,
		DeviceId:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	}); err == nil {
		t.Fatal("SetDriverOnline() accepted inactive driver")
	}
	online, err := rdb.SIsMember(context.Background(), constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember(online) error = %v", err)
	}
	if online {
		t.Fatalf("driver should not be added to %s", constants.RedisDriverOnline)
	}
	if repo.location != nil || repo.updatedStatus != 0 {
		t.Fatalf("driver repository should not be updated: location=%+v status=%d", repo.location, repo.updatedStatus)
	}
}

type setDriverOnlineRepository struct {
	driver        *model.Driver
	location      *model.DriverLocation
	updatedStatus int8
}

type setDriverOnlineCertificationRepository struct {
	certification *model.DriverCertification
}

func (r *setDriverOnlineRepository) Create(context.Context, *model.Driver) error { return nil }

func (r *setDriverOnlineRepository) GetByID(_ context.Context, id uint64) (*model.Driver, error) {
	if r.driver == nil || r.driver.Id != id {
		return nil, errors.New("driver not found")
	}
	return r.driver, nil
}

func (r *setDriverOnlineRepository) GetByPhone(context.Context, string) (*model.Driver, error) {
	return nil, errors.New("not implemented")
}

func (r *setDriverOnlineRepository) List(context.Context, repository.DriverListFilter) ([]*model.Driver, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *setDriverOnlineRepository) ListNearbyDrivers(context.Context, repository.NearbyDriverFilter) ([]*model.DriverLocation, error) {
	return nil, errors.New("not implemented")
}

func (r *setDriverOnlineRepository) UpsertLocation(_ context.Context, location *model.DriverLocation) error {
	r.location = location
	return nil
}

func (r *setDriverOnlineRepository) UpdateLocationStatus(context.Context, uint64, int8) error {
	return nil
}

func (r *setDriverOnlineRepository) UpdateStatusAndLocation(_ context.Context, _ uint64, status int8) error {
	r.updatedStatus = status
	return nil
}

func (r *setDriverOnlineRepository) GetDriverScore(context.Context, uint64) (*model.DriverScore, error) {
	return nil, errors.New("not implemented")
}

func (r *setDriverOnlineRepository) Update(_ context.Context, _ uint64, updates map[string]interface{}) error {
	if v, ok := updates["online_status"].(int8); ok {
		r.updatedStatus = v
	}
	return nil
}

func (r *setDriverOnlineRepository) Delete(context.Context, *model.Driver) error {
	return nil
}

func (r *setDriverOnlineCertificationRepository) Upsert(context.Context, *model.DriverCertification) (*model.DriverCertification, error) {
	return r.certification, nil
}

func (r *setDriverOnlineCertificationRepository) GetByDriverID(_ context.Context, driverID uint64) (*model.DriverCertification, error) {
	if r.certification == nil || r.certification.DriverId != driverID {
		return nil, repository.ErrCertificationNotFound
	}
	return r.certification, nil
}

func (r *setDriverOnlineCertificationRepository) UpdateAudit(context.Context, int64, int64, string, int8) error {
	return nil
}

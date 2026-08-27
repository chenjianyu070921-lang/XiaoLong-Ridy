package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestReportLocationPreservesListenPreference(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{
		DriverRepository: &reportLocationDriverRepository{},
		DriverListenPreferenceRepository: &reportLocationPreferenceRepository{pref: &model.DriverListenPreference{
			DriverID:          25,
			AcceptRealtime:    false,
			AcceptReservation: true,
		}},
		RedisClient: rdb,
		OnlineStore: onlinestore.NewStore(rdb, time.Minute),
	}

	logic := NewReportLocationLogic(context.Background(), svcCtx)
	resp, err := logic.ReportLocation(&proto.ReportLocationRequest{
		DriverId:  25,
		DeviceId:  "device-1",
		Longitude: 116.397,
		Latitude:  39.908,
	})
	if err != nil {
		t.Fatalf("ReportLocation() error = %v", err)
	}
	if resp.GetDriverId() != 25 || resp.GetOnlineStatus() != int32(DriverOnline) || resp.GetKicked() {
		t.Fatalf("ReportLocation() response = %+v", resp)
	}
	realtime, err := rdb.SIsMember(context.Background(), constants.RedisDriverPrefRealtime, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember realtime error = %v", err)
	}
	reservation, err := rdb.SIsMember(context.Background(), constants.RedisDriverPrefReservation, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember reservation error = %v", err)
	}
	if realtime || !reservation {
		t.Fatalf("preference sets realtime=%v reservation=%v, want false/true", realtime, reservation)
	}
	if svcCtx.DriverRepository.(*reportLocationDriverRepository).location == nil {
		t.Fatal("driver_location was not updated")
	}
}

type reportLocationDriverRepository struct {
	location *model.DriverLocation
	updates  []map[string]interface{}
}

func (r *reportLocationDriverRepository) Create(context.Context, *model.Driver) error { return nil }
func (r *reportLocationDriverRepository) GetByID(context.Context, uint64) (*model.Driver, error) {
	return &model.Driver{Id: 25}, nil
}
func (r *reportLocationDriverRepository) GetByPhone(context.Context, string) (*model.Driver, error) {
	return nil, repository.ErrDriverNotFound
}
func (r *reportLocationDriverRepository) List(context.Context, repository.DriverListFilter) ([]*model.Driver, int64, error) {
	return nil, 0, nil
}
func (r *reportLocationDriverRepository) ListNearbyDrivers(context.Context, repository.NearbyDriverFilter) ([]*model.DriverLocation, error) {
	return nil, nil
}
func (r *reportLocationDriverRepository) UpsertLocation(_ context.Context, location *model.DriverLocation) error {
	r.location = location
	return nil
}
func (r *reportLocationDriverRepository) UpdateLocationStatus(context.Context, uint64, int8) error {
	return nil
}
func (r *reportLocationDriverRepository) GetDriverScore(context.Context, uint64) (*model.DriverScore, error) {
	return nil, nil
}
func (r *reportLocationDriverRepository) Update(_ context.Context, _ uint64, updates map[string]interface{}) error {
	r.updates = append(r.updates, updates)
	return nil
}
func (r *reportLocationDriverRepository) Delete(context.Context, *model.Driver) error { return nil }
func (r *reportLocationDriverRepository) UpdateStatusAndLocation(context.Context, uint64, int8) error {
	return nil
}

type reportLocationPreferenceRepository struct {
	pref *model.DriverListenPreference
}

func (r *reportLocationPreferenceRepository) GetByDriverID(context.Context, uint64) (*model.DriverListenPreference, error) {
	return r.pref, nil
}
func (r *reportLocationPreferenceRepository) Upsert(context.Context, *model.DriverListenPreference) error {
	return nil
}

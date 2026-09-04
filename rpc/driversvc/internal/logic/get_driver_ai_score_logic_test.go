package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

func TestGetDriverAiScoreReadsStoredMetricsBeforeScoring(t *testing.T) {
	repo := &aiScoreDriverRepository{
		score: &model.DriverScore{
			DriverId:            25,
			Score:               90,
			Level:               3,
			MonthOrders:         20,
			MonthCancelRate:     25,
			MonthComplaintCount: 1,
		},
	}
	logic := NewGetDriverAiScoreLogic(context.Background(), &svc.ServiceContext{DriverRepository: repo})

	resp, err := logic.GetDriverAiScore(&proto.GetDriverAiScoreRequest{DriverId: 25})
	if err != nil {
		t.Fatalf("GetDriverAiScore() error = %v", err)
	}
	if !repo.getScoreCalled || repo.getScoreDriverID != 25 {
		t.Fatalf("GetDriverScore() called=%v driverID=%d", repo.getScoreCalled, repo.getScoreDriverID)
	}
	if repo.refreshCalled {
		t.Fatal("RefreshDriverScoreMetrics() should not be called by GetDriverAiScore")
	}
	if resp.GetDriverId() != 25 || resp.GetAiScore() != 77.3 || resp.GetLevel() != 3 || resp.GetDegraded() {
		t.Fatalf("GetDriverAiScore() response = %+v", resp)
	}
	if len(resp.GetFactors()) != 4 {
		t.Fatalf("factor count = %d, want 4", len(resp.GetFactors()))
	}
	wantKeys := []string{"service_score", "cancel_rate", "complaint", "month_orders"}
	for i, key := range wantKeys {
		if resp.GetFactors()[i].GetKey() != key {
			t.Fatalf("factor[%d].key = %q, want %q", i, resp.GetFactors()[i].GetKey(), key)
		}
	}
}

func TestGetDriverAiScoreDegradesWhenScoreLookupFails(t *testing.T) {
	repo := &aiScoreDriverRepository{scoreErr: errors.New("metrics db unavailable")}
	logic := NewGetDriverAiScoreLogic(context.Background(), &svc.ServiceContext{DriverRepository: repo})

	resp, err := logic.GetDriverAiScore(&proto.GetDriverAiScoreRequest{DriverId: 25})
	if err != nil {
		t.Fatalf("GetDriverAiScore() error = %v", err)
	}
	if !resp.GetDegraded() || resp.GetDriverId() != 25 || resp.GetAiScore() != 0 || resp.GetDegradeReason() == "" {
		t.Fatalf("GetDriverAiScore() response = %+v", resp)
	}
}

type aiScoreDriverRepository struct {
	getScoreCalled   bool
	getScoreDriverID uint64
	refreshCalled    bool
	refreshDriverID  uint64
	refreshStart     time.Time
	refreshEnd       time.Time
	score            *model.DriverScore
	scoreErr         error
}

func (r *aiScoreDriverRepository) Create(context.Context, *model.Driver) error {
	return nil
}

func (r *aiScoreDriverRepository) GetByID(context.Context, uint64) (*model.Driver, error) {
	return nil, repository.ErrDriverNotFound
}

func (r *aiScoreDriverRepository) GetByPhone(context.Context, string) (*model.Driver, error) {
	return nil, repository.ErrDriverNotFound
}

func (r *aiScoreDriverRepository) List(context.Context, repository.DriverListFilter) ([]*model.Driver, int64, error) {
	return nil, 0, nil
}

func (r *aiScoreDriverRepository) ListNearbyDrivers(context.Context, repository.NearbyDriverFilter) ([]*model.DriverLocation, error) {
	return nil, nil
}

func (r *aiScoreDriverRepository) UpsertLocation(context.Context, *model.DriverLocation) error {
	return nil
}

func (r *aiScoreDriverRepository) UpdateLocationStatus(context.Context, uint64, int8) error {
	return nil
}

func (r *aiScoreDriverRepository) UpdateStatusAndLocation(context.Context, uint64, int8) error {
	return nil
}

func (r *aiScoreDriverRepository) RefreshDriverScoreMetrics(_ context.Context, driverID uint64, startAt, endAt time.Time) (*model.DriverScore, error) {
	r.refreshCalled = true
	r.refreshDriverID = driverID
	r.refreshStart = startAt
	r.refreshEnd = endAt
	return r.score, r.scoreErr
}

func (r *aiScoreDriverRepository) GetDriverScore(_ context.Context, driverID uint64) (*model.DriverScore, error) {
	r.getScoreCalled = true
	r.getScoreDriverID = driverID
	return r.score, r.scoreErr
}

func (r *aiScoreDriverRepository) Update(context.Context, uint64, map[string]interface{}) error {
	return nil
}

func (r *aiScoreDriverRepository) Delete(context.Context, *model.Driver) error {
	return nil
}

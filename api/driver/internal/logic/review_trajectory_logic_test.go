package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

type fakeReviewRepository struct {
	driverID int64
	page     int32
	pageSize int32
}

func (f *fakeReviewRepository) ListByDriver(_ context.Context, driverID int64, page, pageSize int32) ([]svc.PassengerReviewRecord, int64, error) {
	f.driverID = driverID
	f.page = page
	f.pageSize = pageSize
	return []svc.PassengerReviewRecord{{
		OrderID:   1001,
		Rating:    5,
		Comment:   "准时",
		CreatedAt: time.Unix(123, 0),
	}}, 1, nil
}

type fakeTrajectoryRepository struct {
	driverID int64
	orderID  int64
	recorded *svc.TrajectoryRecord
}

func (f *fakeTrajectoryRepository) ListByOrder(_ context.Context, driverID, orderID int64) ([]svc.TrajectoryRecord, error) {
	f.driverID = driverID
	f.orderID = orderID
	return []svc.TrajectoryRecord{{
		Longitude:  116.397,
		Latitude:   39.908,
		SpeedKmh:   35.5,
		Heading:    90,
		RecordedAt: time.Unix(456, 0),
	}}, nil
}

func (f *fakeTrajectoryRepository) RecordPoint(_ context.Context, record *svc.TrajectoryRecord) error {
	copied := *record
	f.recorded = &copied
	return nil
}

func TestListPassengerReviewsReadsRealStorage(t *testing.T) {
	repo := &fakeReviewRepository{}
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{ReviewRepository: repo})

	resp, err := logic.ListPassengerReviews(25, &types.ListPassengerReviewsRequest{Page: 0, PageSize: 200})
	if err != nil {
		t.Fatalf("ListPassengerReviews() error = %v", err)
	}
	if repo.driverID != 25 || repo.page != 1 || repo.pageSize != 100 {
		t.Fatalf("repository request driver=%d page=%d pageSize=%d", repo.driverID, repo.page, repo.pageSize)
	}
	if resp.Total != 1 || resp.Page != 1 || resp.PageSize != 100 || resp.Degraded || resp.Message != "" {
		t.Fatalf("ListPassengerReviews() response metadata = %+v", resp)
	}
	if len(resp.List) != 1 || resp.List[0].OrderID != 1001 || resp.List[0].Rating != 5 || resp.List[0].Comment != "准时" || resp.List[0].CreatedAt != 123 {
		t.Fatalf("ListPassengerReviews() list = %+v", resp.List)
	}
}

func TestGetTripTrajectoryReadsRealStorage(t *testing.T) {
	repo := &fakeTrajectoryRepository{}
	logic := NewTrajectoryLogic(context.Background(), &svc.ServiceContext{TrajectoryRepository: repo})

	resp, err := logic.GetTripTrajectory(25, &types.GetTripTrajectoryRequest{OrderID: 1001})
	if err != nil {
		t.Fatalf("GetTripTrajectory() error = %v", err)
	}
	if repo.driverID != 25 || repo.orderID != 1001 {
		t.Fatalf("repository request driver=%d order=%d", repo.driverID, repo.orderID)
	}
	if resp.OrderID != 1001 || resp.Degraded || resp.Message != "" {
		t.Fatalf("GetTripTrajectory() response metadata = %+v", resp)
	}
	if len(resp.Points) != 1 || resp.Points[0].Longitude != 116.397 || resp.Points[0].Latitude != 39.908 || resp.Points[0].SpeedKmh != 35.5 || resp.Points[0].Heading != 90 || resp.Points[0].CreatedAt != 456 {
		t.Fatalf("GetTripTrajectory() points = %+v", resp.Points)
	}
}

func TestListPassengerReviewsReturnsStorageUnavailableError(t *testing.T) {
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{})

	resp, err := logic.ListPassengerReviews(25, &types.ListPassengerReviewsRequest{Page: 1, PageSize: 20})
	if !errors.Is(err, ErrReviewStorageNotConfigured) {
		t.Fatalf("ListPassengerReviews() error = %v, want %v", err, ErrReviewStorageNotConfigured)
	}
	if resp != nil {
		t.Fatalf("ListPassengerReviews() response = %+v, want nil", resp)
	}
}

func TestGetTripTrajectoryReturnsStorageUnavailableError(t *testing.T) {
	logic := NewTrajectoryLogic(context.Background(), &svc.ServiceContext{})

	resp, err := logic.GetTripTrajectory(25, &types.GetTripTrajectoryRequest{OrderID: 1001})
	if !errors.Is(err, ErrTrajectoryStorageNotConfigured) {
		t.Fatalf("GetTripTrajectory() error = %v, want %v", err, ErrTrajectoryStorageNotConfigured)
	}
	if resp != nil {
		t.Fatalf("GetTripTrajectory() response = %+v, want nil", resp)
	}
}

package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

type fakeReviewRepository struct {
	receivedList  []svc.PassengerReview
	receivedTotal int64
}

func (f *fakeReviewRepository) ListPassengerReviewsByDriver(_ context.Context, driverID int64, page, pageSize int32) ([]svc.PassengerReview, int64, error) {
	return f.receivedList, f.receivedTotal, nil
}

func (f *fakeReviewRepository) CountAndAvgRatingByDriver(_ context.Context, driverID int64) (int64, float64, error) {
	var count int64
	var sum float64
	for _, r := range f.receivedList {
		if int64(r.DriverID) == driverID {
			count++
			sum += float64(r.Rating)
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return count, sum / float64(count), nil
}

func TestListReceivedReviewsReturnsPassengerReviewsForCurrentDriver(t *testing.T) {
	repo := &fakeReviewRepository{
		receivedList: []svc.PassengerReview{{
			ID:        501,
			OrderID:   1001,
			UserID:    300,
			DriverID:  25,
			Rating:    5,
			Comment:   "准时专业",
			Tags:      "准时,车内整洁",
			CreatedAt: 123,
		}},
		receivedTotal: 1,
	}
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{ReviewRepository: repo})

	resp, err := logic.ListReceivedReviews(25, &types.ListReviewsRequest{Page: 0, PageSize: 1000})
	if err != nil {
		t.Fatalf("ListReceivedReviews() error = %v", err)
	}
	if resp.Total != 1 || resp.Page != 1 || resp.PageSize != 100 || len(resp.List) != 1 {
		t.Fatalf("ListReceivedReviews() response = %+v", resp)
	}
	item := resp.List[0]
	if item.Direction != "received" || item.OrderID != 1001 || item.UserID != 0 ||
		item.DriverID != 25 || item.Rating != 5 || item.Comment != "准时专业" || item.Tags != "准时,车内整洁" {
		t.Fatalf("ListReceivedReviews() item = %+v", item)
	}
}

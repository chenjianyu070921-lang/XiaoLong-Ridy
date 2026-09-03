package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

type fakeReviewRepository struct {
	receivedList  []svc.PassengerReview
	receivedTotal int64
	givenList     []svc.DriverOrderReview
	givenTotal    int64
	created       *svc.DriverOrderReview
	createErr     error
}

func (f *fakeReviewRepository) ListPassengerReviewsByDriver(_ context.Context, driverID int64, page, pageSize int32) ([]svc.PassengerReview, int64, error) {
	return f.receivedList, f.receivedTotal, nil
}

func (f *fakeReviewRepository) CreateDriverReview(_ context.Context, review *svc.DriverOrderReview) error {
	copied := *review
	if copied.ID == 0 {
		copied.ID = 7001
	}
	f.created = &copied
	*review = copied
	return f.createErr
}

func (f *fakeReviewRepository) ListDriverReviewsByDriver(_ context.Context, driverID int64, page, pageSize int32) ([]svc.DriverOrderReview, int64, error) {
	return f.givenList, f.givenTotal, nil
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

func TestSubmitDriverReviewCreatesReviewForCompletedOwnedOrder(t *testing.T) {
	repo := &fakeReviewRepository{}
	orderClient := &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_COMPLETED}
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{
		OrderClient:       orderClient,
		ReviewRepository: repo,
	})

	resp, err := logic.SubmitDriverReview(25, &types.SubmitDriverReviewRequest{
		OrderID: 1001,
		Rating:  4,
		Comment: "  乘客礼貌  ",
		Tags:    " 准时上车 ",
	})
	if err != nil {
		t.Fatalf("SubmitDriverReview() error = %v", err)
	}
	if resp.ReviewID != 7001 || resp.OrderID != 1001 || resp.DriverID != 25 || resp.Rating != 4 {
		t.Fatalf("SubmitDriverReview() response = %+v", resp)
	}
	if repo.created == nil {
		t.Fatal("SubmitDriverReview() did not create review")
	}
	if repo.created.UserID != 300 || repo.created.Comment != "乘客礼貌" || repo.created.Tags != "准时上车" {
		t.Fatalf("created review = %+v", repo.created)
	}
}

func TestSubmitDriverReviewRejectsOtherDriverOrder(t *testing.T) {
	repo := &fakeReviewRepository{}
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{
		OrderClient:       &fakeOrderClient{getOrderResponseDriverID: 26, getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_COMPLETED},
		ReviewRepository: repo,
	})

	_, err := logic.SubmitDriverReview(25, &types.SubmitDriverReviewRequest{OrderID: 1001, Rating: 5})
	if err != ErrForbiddenDriverResource {
		t.Fatalf("SubmitDriverReview() error = %v, want %v", err, ErrForbiddenDriverResource)
	}
	if repo.created != nil {
		t.Fatalf("SubmitDriverReview() should not create review for another driver's order: %+v", repo.created)
	}
}

func TestSubmitDriverReviewRequiresCompletedOrder(t *testing.T) {
	repo := &fakeReviewRepository{}
	logic := NewReviewLogic(context.Background(), &svc.ServiceContext{
		OrderClient:       &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY},
		ReviewRepository: repo,
	})

	_, err := logic.SubmitDriverReview(25, &types.SubmitDriverReviewRequest{OrderID: 1001, Rating: 5})
	if err != ErrInvalidParam {
		t.Fatalf("SubmitDriverReview() error = %v, want %v", err, ErrInvalidParam)
	}
	if repo.created != nil {
		t.Fatalf("SubmitDriverReview() should not create review for unfinished order: %+v", repo.created)
	}
}

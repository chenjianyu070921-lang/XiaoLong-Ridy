package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func seedTimeoutOrder(t *testing.T, repo *repository.MemoryOrderRepository, userID uint64, status int8, createdAt time.Time) *model.RideOrder {
	t.Helper()
	order := &model.RideOrder{
		OrderNo:       nextTestOrderNo(),
		UserId:        userID,
		CarType:       1,
		FromAddress:   "起点",
		FromLongitude: 116.1,
		FromLatitude:  39.1,
		ToAddress:     "终点",
		ToLongitude:   116.2,
		ToLatitude:    39.2,
		Status:        status,
		CreatedAt:     createdAt,
	}
	statusLog := &model.OrderStatusLog{
		FromStatus:   0,
		ToStatus:     status,
		OperatorType: "user",
		OperatorId:   userID,
	}
	if err := repo.Create(context.Background(), order, statusLog); err != nil {
		t.Fatalf("seed timeout orderclient error = %v", err)
	}
	return order
}

func TestListTimeoutOrdersReturnsOnlyExpiredWaitAccept(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	now := time.Now()
	expired := seedTimeoutOrder(t, repo, 1001, 1, now.Add(-10*time.Minute))
	seedTimeoutOrder(t, repo, 1001, 1, now.Add(-time.Minute))
	seedTimeoutOrder(t, repo, 1001, 2, now.Add(-10*time.Minute))
	seedTimeoutOrder(t, repo, 1001, 6, now.Add(-10*time.Minute))
	l := NewListTimeoutOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ListTimeoutOrders(&proto.ListTimeoutOrdersRequest{TimeoutSeconds: 300})
	if err != nil {
		t.Fatalf("ListTimeoutOrders() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("ListTimeoutOrders() response = %+v", resp)
	}
	if resp.List[0].OrderId != int64(expired.Id) {
		t.Fatalf("ListTimeoutOrders() orderclient = %+v, want id %d", resp.List[0], expired.Id)
	}
}

func TestListTimeoutOrdersDefaultTimeout(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	now := time.Now()
	old := seedTimeoutOrder(t, repo, 1001, 1, now.Add(-6*time.Minute))
	seedTimeoutOrder(t, repo, 1001, 1, now.Add(-4*time.Minute))
	l := NewListTimeoutOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ListTimeoutOrders(&proto.ListTimeoutOrdersRequest{})
	if err != nil {
		t.Fatalf("ListTimeoutOrders() error = %v", err)
	}
	if resp.Total != 1 || resp.List[0].OrderId != int64(old.Id) {
		t.Fatalf("ListTimeoutOrders() response = %+v", resp)
	}
}

func TestListTimeoutOrdersPaginationOldestFirst(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	now := time.Now()
	oldest := seedTimeoutOrder(t, repo, 1001, 1, now.Add(-30*time.Minute))
	seedTimeoutOrder(t, repo, 1001, 1, now.Add(-20*time.Minute))
	seedTimeoutOrder(t, repo, 1001, 1, now.Add(-10*time.Minute))
	l := NewListTimeoutOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	first, err := l.ListTimeoutOrders(&proto.ListTimeoutOrdersRequest{TimeoutSeconds: 300, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListTimeoutOrders() error = %v", err)
	}
	if first.Total != 3 || len(first.List) != 2 || first.List[0].OrderId != int64(oldest.Id) {
		t.Fatalf("ListTimeoutOrders() first page = %+v", first)
	}

	second, err := l.ListTimeoutOrders(&proto.ListTimeoutOrdersRequest{TimeoutSeconds: 300, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListTimeoutOrders() error = %v", err)
	}
	if second.Total != 3 || len(second.List) != 1 {
		t.Fatalf("ListTimeoutOrders() second page = %+v", second)
	}
}

func TestListTimeoutOrdersRejectsNegativeTimeout(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewListTimeoutOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ListTimeoutOrders(&proto.ListTimeoutOrdersRequest{TimeoutSeconds: -1})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ListTimeoutOrders() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

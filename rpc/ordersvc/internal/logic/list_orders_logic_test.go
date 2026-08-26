package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestListOrdersFiltersAndPagination(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	seedOrder(t, repo, 1001, 0, 1)
	seedOrder(t, repo, 1001, 2002, 2)
	seedOrder(t, repo, 1001, 0, 1)
	seedOrder(t, repo, 2002, 0, 1)
	l := NewListOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ListOrders(&proto.ListOrdersRequest{
		UserId:   1001,
		Status:   proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 1 || resp.Page != 1 || resp.PageSize != 1 {
		t.Fatalf("ListOrders() response = %+v", resp)
	}
	if resp.List[0].Status != proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT {
		t.Fatalf("ListOrders() item = %+v", resp.List[0])
	}
	if got := resp.List[0]; got.UserId != 1001 || got.CarType != 1 || got.FromLongitude == 0 || got.UpdatedAt == 0 {
		t.Fatalf("ListOrders() full item fields = %+v", got)
	}
}

func TestListOrdersDefaults(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	seedOrder(t, repo, 1001, 0, 1)
	seedOrder(t, repo, 1001, 2002, 2)
	seedOrder(t, repo, 2002, 0, 3)
	l := NewListOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ListOrders(&proto.ListOrdersRequest{})
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if resp.Total != 3 || len(resp.List) != 3 || resp.Page != 1 || resp.PageSize != 20 {
		t.Fatalf("ListOrders() response = %+v", resp)
	}
}

func TestListOrdersRejectsInvalidStatus(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewListOrdersLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ListOrders(&proto.ListOrdersRequest{Status: proto.OrderStatus(7)})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ListOrders() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

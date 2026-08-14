package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestStartTripSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewStartTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.StartTrip(&proto.StartTripRequest{
		OrderId:  int64(order.Id),
		DriverId: 2002,
	})
	if err != nil {
		t.Fatalf("StartTrip() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_ON_TRIP {
		t.Fatalf("StartTrip() status = %v, want ON_TRIP", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 3 {
		t.Fatalf("started order status = %d, want 3", fresh.Status)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 2 || logs[1].ToStatus != 3 || logs[1].OperatorType != "driver" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestStartTripRejectWrongStatus(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewStartTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.StartTrip(&proto.StartTripRequest{
		OrderId:  int64(order.Id),
		DriverId: 2002,
	})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("StartTrip() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestStartTripRejectDriverMismatch(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewStartTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.StartTrip(&proto.StartTripRequest{
		OrderId:  int64(order.Id),
		DriverId: 8888,
	})
	if !errors.Is(err, ErrDriverNotMatched) {
		t.Fatalf("StartTrip() error = %v, want %v", err, ErrDriverNotMatched)
	}
}

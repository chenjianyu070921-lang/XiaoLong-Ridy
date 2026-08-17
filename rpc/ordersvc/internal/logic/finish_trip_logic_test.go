package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestFinishTripSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualDistanceM:  15000,
		ActualDurationS:  2400,
		ActualPriceCents: 5200,
	})
	if err != nil {
		t.Fatalf("FinishTrip() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_WAIT_PAY {
		t.Fatalf("FinishTrip() status = %v, want WAIT_PAY", resp.Status)
	}
	if resp.PayableAmountCents != 5200 {
		t.Fatalf("FinishTrip() payable = %d, want 5200", resp.PayableAmountCents)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 4 {
		t.Fatalf("finished order status = %d, want 4", fresh.Status)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 3 || logs[1].ToStatus != 4 {
		t.Fatalf("status logs = %+v", logs)
	}
	if !strings.Contains(logs[1].Remark, "实际费用=5200分") {
		t.Fatalf("finish remark = %q, want actual fee snapshot", logs[1].Remark)
	}
}

func TestFinishTripRejectWrongStatus(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualPriceCents: 1000,
	})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("FinishTrip() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestFinishTripRejectNegativePrice(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualPriceCents: -1,
	})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("FinishTrip() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

func TestFinishTripRejectDriverMismatch(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         8888,
		ActualPriceCents: 1000,
	})
	if !errors.Is(err, ErrDriverNotMatched) {
		t.Fatalf("FinishTrip() error = %v, want %v", err, ErrDriverNotMatched)
	}
}

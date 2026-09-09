package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestConfirmArriveSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewConfirmArriveLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ConfirmArrive(&proto.ConfirmArriveRequest{
		OrderId:  int64(order.Id),
		DriverId: 2002,
	})
	if err != nil {
		t.Fatalf("ConfirmArrive() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_ACCEPTED {
		t.Fatalf("ConfirmArrive() status = %v, want ACCEPTED", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 2 {
		t.Fatalf("arrive orderclient status = %d, want 2", fresh.Status)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 2 || logs[1].ToStatus != 2 || logs[1].Remark != "司机已到达" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestConfirmArriveRejectWrongStatus(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewConfirmArriveLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmArrive(&proto.ConfirmArriveRequest{
		OrderId:  int64(order.Id),
		DriverId: 2002,
	})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("ConfirmArrive() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestConfirmArriveRejectDriverMismatch(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewConfirmArriveLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmArrive(&proto.ConfirmArriveRequest{
		OrderId:  int64(order.Id),
		DriverId: 8888,
	})
	if !errors.Is(err, ErrDriverNotMatched) {
		t.Fatalf("ConfirmArrive() error = %v, want %v", err, ErrDriverNotMatched)
	}
}

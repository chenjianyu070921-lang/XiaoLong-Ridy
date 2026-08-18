package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestAcceptOrderSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewAcceptOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.AcceptOrder(&proto.AcceptOrderRequest{
		OrderId:  int64(order.Id),
		DriverId: 2002,
	})
	if err != nil {
		t.Fatalf("AcceptOrder() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_ACCEPTED {
		t.Fatalf("AcceptOrder() status = %v, want ACCEPTED", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 2 || fresh.DriverId != 2002 {
		t.Fatalf("accepted orderclient = %+v", fresh)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 1 || logs[1].ToStatus != 2 || logs[1].OperatorType != "driver" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestAcceptOrderRejectAlreadyAccepted(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewAcceptOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.AcceptOrder(&proto.AcceptOrderRequest{
		OrderId:  int64(order.Id),
		DriverId: 2003,
	})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("AcceptOrder() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestAcceptOrderRejectInvalidParams(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewAcceptOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	if _, err := l.AcceptOrder(&proto.AcceptOrderRequest{OrderId: 0, DriverId: 2002}); !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("AcceptOrder() error = %v, want %v", err, ErrInvalidOrderParams)
	}
	if _, err := l.AcceptOrder(&proto.AcceptOrderRequest{OrderId: 1, DriverId: 0}); !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("AcceptOrder() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

func TestAcceptOrderNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewAcceptOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.AcceptOrder(&proto.AcceptOrderRequest{OrderId: 999, DriverId: 2002})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("AcceptOrder() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}

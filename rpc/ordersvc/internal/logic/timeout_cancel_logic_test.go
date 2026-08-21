package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestTimeoutCancelSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewTimeoutCancelLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.TimeoutCancel(&proto.TimeoutCancelRequest{
		OrderId: int64(order.Id),
		Reason:  "超时未接单",
	})
	if err != nil {
		t.Fatalf("TimeoutCancel() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Fatalf("TimeoutCancel() status = %v, want CANCELLED", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 6 || fresh.CancelBy != "system" || fresh.CancelReason != "超时未接单" {
		t.Fatalf("timeout cancelled orderclient = %+v", fresh)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 1 || logs[1].ToStatus != 6 || logs[1].OperatorType != "system" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestTimeoutCancelDefaultReason(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewTimeoutCancelLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	if _, err := l.TimeoutCancel(&proto.TimeoutCancelRequest{OrderId: int64(order.Id), Reason: "  "}); err != nil {
		t.Fatalf("TimeoutCancel() error = %v", err)
	}
	logs := repo.StatusLogs(order.Id)
	if logs[1].Remark != "超时未接单" {
		t.Fatalf("timeout remark = %q, want default", logs[1].Remark)
	}
}

func TestTimeoutCancelRejectOnTrip(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	l := NewTimeoutCancelLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.TimeoutCancel(&proto.TimeoutCancelRequest{
		OrderId: int64(order.Id),
		Reason:  "超时",
	})
	if !errors.Is(err, ErrOrderStatusNotCancelable) {
		t.Fatalf("TimeoutCancel() error = %v, want %v", err, ErrOrderStatusNotCancelable)
	}
}

func TestTimeoutCancelNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewTimeoutCancelLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.TimeoutCancel(&proto.TimeoutCancelRequest{OrderId: 999, Reason: "超时"})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("TimeoutCancel() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}

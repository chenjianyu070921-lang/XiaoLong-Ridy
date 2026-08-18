package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestConfirmPaidSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 4) // 待支付
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{OrderId: int64(order.Id), PaymentNo: "PAY123"})
	if err != nil {
		t.Fatalf("ConfirmPaid() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_COMPLETED {
		t.Fatalf("ConfirmPaid() status = %v, want COMPLETED", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != constants.OrderStatusCompleted {
		t.Fatalf("completed orderclient status = %d, want %d", fresh.Status, constants.OrderStatusCompleted)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != constants.OrderStatusWaitPay || logs[1].ToStatus != constants.OrderStatusCompleted {
		t.Fatalf("status logs = %+v", logs)
	}
	if !strings.Contains(logs[1].Remark, "PAY123") {
		t.Fatalf("confirm remark = %q, want payment no", logs[1].Remark)
	}
}

func TestConfirmPaidRejectsInvalidParams(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{OrderId: 0})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ConfirmPaid() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

func TestConfirmPaidRejectsNonWaitPay(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3) // 行程中
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{OrderId: int64(order.Id)})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("ConfirmPaid() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestConfirmPaidOrderNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{OrderId: 99999})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("ConfirmPaid() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}

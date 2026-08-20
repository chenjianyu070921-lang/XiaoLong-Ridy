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
	pay "XiaoLong-Ridy/rpc/paysvc/pay"
)

func TestConfirmPaidSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 4) // 待支付
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PayClient:       &fakePayClient{},
	})

	resp, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
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

	cases := []*proto.ConfirmPaidRequest{
		{OrderId: 0, PaymentNo: "PAY123", AmountCents: 5200, PaidAt: 1787038459},
		{OrderId: 1, PaymentNo: "", AmountCents: 5200, PaidAt: 1787038459},
		{OrderId: 1, PaymentNo: "PAY123", AmountCents: 0, PaidAt: 1787038459},
		{OrderId: 1, PaymentNo: "PAY123", AmountCents: 5200, PaidAt: 0},
	}
	for _, in := range cases {
		if _, err := l.ConfirmPaid(in); !errors.Is(err, ErrInvalidOrderParams) {
			t.Fatalf("ConfirmPaid(%+v) error = %v, want %v", in, err, ErrInvalidOrderParams)
		}
	}
}

func TestConfirmPaidRejectsNonWaitPay(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3) // 行程中
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if !errors.Is(err, ErrOrderStatusNotAllowed) {
		t.Fatalf("ConfirmPaid() error = %v, want %v", err, ErrOrderStatusNotAllowed)
	}
}

func TestConfirmPaidOrderNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      99999,
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("ConfirmPaid() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}

func TestConfirmPaidPayClientNotConfigured(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 4)
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if err == nil || !strings.Contains(err.Error(), "pay client not configured") {
		t.Fatalf("ConfirmPaid() error = %v, want pay client not configured", err)
	}
}

func TestConfirmPaidRejectsPaymentMismatch(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 4)

	// 支付单号不匹配
	pc := &fakePayClient{getPayment: func(_ context.Context, _ *pay.GetPaymentRequest) (*pay.GetPaymentResponse, error) {
		return &pay.GetPaymentResponse{
			PaymentNo:   "PAY_OTHER",
			OrderId:     int64(order.Id),
			AmountCents: 5200,
			Status:      2,
		}, nil
	}}
	l := NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo, PayClient: pc})
	_, err := l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if err == nil || !strings.Contains(err.Error(), "payment verification failed") {
		t.Fatalf("ConfirmPaid() error = %v, want payment verification failed", err)
	}

	// 订单不匹配
	pc = &fakePayClient{getPayment: func(_ context.Context, _ *pay.GetPaymentRequest) (*pay.GetPaymentResponse, error) {
		return &pay.GetPaymentResponse{
			PaymentNo:   "PAY123",
			OrderId:     int64(order.Id) + 1,
			AmountCents: 5200,
			Status:      2,
		}, nil
	}}
	l = NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo, PayClient: pc})
	_, err = l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if err == nil || !strings.Contains(err.Error(), "payment verification failed") {
		t.Fatalf("ConfirmPaid() error = %v, want payment verification failed", err)
	}

	// 金额不匹配
	pc = &fakePayClient{getPayment: func(_ context.Context, _ *pay.GetPaymentRequest) (*pay.GetPaymentResponse, error) {
		return &pay.GetPaymentResponse{
			PaymentNo:   "PAY123",
			OrderId:     int64(order.Id),
			AmountCents: 9999,
			Status:      2,
		}, nil
	}}
	l = NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo, PayClient: pc})
	_, err = l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if err == nil || !strings.Contains(err.Error(), "payment verification failed") {
		t.Fatalf("ConfirmPaid() error = %v, want payment verification failed", err)
	}

	// 支付状态非成功
	pc = &fakePayClient{getPayment: func(_ context.Context, _ *pay.GetPaymentRequest) (*pay.GetPaymentResponse, error) {
		return &pay.GetPaymentResponse{
			PaymentNo:   "PAY123",
			OrderId:     int64(order.Id),
			AmountCents: 5200,
			Status:      1, // 待支付
		}, nil
	}}
	l = NewConfirmPaidLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo, PayClient: pc})
	_, err = l.ConfirmPaid(&proto.ConfirmPaidRequest{
		OrderId:      int64(order.Id),
		PaymentNo:    "PAY123",
		AmountCents:  5200,
		PaidAt:       1787038459,
	})
	if err == nil || !strings.Contains(err.Error(), "payment verification failed") {
		t.Fatalf("ConfirmPaid() error = %v, want payment verification failed", err)
	}
}

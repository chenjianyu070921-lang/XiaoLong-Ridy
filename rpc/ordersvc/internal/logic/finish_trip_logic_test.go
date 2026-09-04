package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/config"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

	"google.golang.org/grpc"
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
		t.Fatalf("finished orderclient status = %d, want 4", fresh.Status)
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

type fakePayClient struct {
	createPayment  func(ctx context.Context, in *pay.CreatePaymentRequest) (*pay.CreatePaymentResponse, error)
	getPayment     func(ctx context.Context, in *pay.GetPaymentRequest) (*pay.GetPaymentResponse, error)
	gotOrderId     int64
	gotUserId      int64
	gotAmountCents int64
	gotChannel     payproto.PayChannel
}

func (f *fakePayClient) CreatePayment(_ context.Context, in *pay.CreatePaymentRequest, _ ...grpc.CallOption) (*pay.CreatePaymentResponse, error) {
	f.gotOrderId = in.OrderId
	f.gotUserId = in.UserId
	f.gotAmountCents = in.AmountCents
	f.gotChannel = in.Channel
	if f.createPayment != nil {
		return f.createPayment(nil, in)
	}
	return &pay.CreatePaymentResponse{PaymentNo: "PAY202608170001"}, nil
}

func (f *fakePayClient) GetPayment(_ context.Context, in *pay.GetPaymentRequest, _ ...grpc.CallOption) (*pay.GetPaymentResponse, error) {
	if f.getPayment != nil {
		return f.getPayment(nil, in)
	}
	return &pay.GetPaymentResponse{
		PaymentNo:   in.PaymentNo,
		OrderId:     in.OrderId,
		AmountCents: 5200,
		Status:      2,
	}, nil
}

func (f *fakePayClient) NotifyPayment(_ context.Context, _ *pay.NotifyPaymentRequest, _ ...grpc.CallOption) (*pay.NotifyPaymentResponse, error) {
	return &pay.NotifyPaymentResponse{Success: true}, nil
}

func (f *fakePayClient) RefundPayment(_ context.Context, _ *pay.RefundPaymentRequest, _ ...grpc.CallOption) (*pay.RefundPaymentResponse, error) {
	return &pay.RefundPaymentResponse{}, nil
}

func (f *fakePayClient) SettleOrder(_ context.Context, _ *pay.SettleOrderRequest, _ ...grpc.CallOption) (*pay.SettleOrderResponse, error) {
	return &pay.SettleOrderResponse{}, nil
}

func (f *fakePayClient) GetSettlement(_ context.Context, _ *pay.GetSettlementRequest, _ ...grpc.CallOption) (*pay.GetSettlementResponse, error) {
	return &pay.GetSettlementResponse{}, nil
}

func (f *fakePayClient) ListSettlements(_ context.Context, _ *pay.ListSettlementsRequest, _ ...grpc.CallOption) (*pay.ListSettlementsResponse, error) {
	return &pay.ListSettlementsResponse{}, nil
}

func TestFinishTripCreatesPayment(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	pc := &fakePayClient{}
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PayClient:       pc,
		Config:          config.Config{PayChannel: 2},
	})

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
	if pc.gotOrderId != int64(order.Id) {
		t.Fatalf("payment orderclient id = %d, want %d", pc.gotOrderId, order.Id)
	}
	if pc.gotUserId != 1001 {
		t.Fatalf("payment user id = %d, want 1001", pc.gotUserId)
	}
	if pc.gotAmountCents != 5200 {
		t.Fatalf("payment amount = %d, want 5200", pc.gotAmountCents)
	}
	if pc.gotChannel != 2 {
		t.Fatalf("payment channel = %v, want 2", pc.gotChannel)
	}
}

func TestFinishTripRejectPriceMismatch(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	pc := &fakePriceClient{estimatePrice: func(_ context.Context, _ *price.EstimatePriceRequest) (*price.EstimatePriceResponse, error) {
		return &price.EstimatePriceResponse{TotalCents: 5000}, nil
	}}
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PriceClient:     pc,
	})

	// 服务端计价 5000 分，司机上报 6000 分偏差 20% > 10%，拒绝。
	_, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualDistanceM:  15000,
		ActualDurationS:  2400,
		ActualPriceCents: 6000,
	})
	if !errors.Is(err, ErrPriceMismatch) {
		t.Fatalf("FinishTrip() error = %v, want %v", err, ErrPriceMismatch)
	}
}

func TestFinishTripServerPriceAuthority(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	pc := &fakePriceClient{estimatePrice: func(_ context.Context, _ *price.EstimatePriceRequest) (*price.EstimatePriceResponse, error) {
		return &price.EstimatePriceResponse{TotalCents: 5600}, nil
	}}
	payc := &fakePayClient{}
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PriceClient:     pc,
		PayClient:       payc,
	})

	// 服务端计价 5600 分，司机上报 5800 分偏差 3.6% < 10%，通过且支付单金额取服务端值。
	resp, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualDistanceM:  15000,
		ActualDurationS:  2400,
		ActualPriceCents: 5800,
	})
	if err != nil {
		t.Fatalf("FinishTrip() error = %v", err)
	}
	if resp.PayableAmountCents != 5600 {
		t.Fatalf("FinishTrip() payable = %d, want 5600", resp.PayableAmountCents)
	}
	if payc.gotAmountCents != 5600 {
		t.Fatalf("payment amount = %d, want 5600", payc.gotAmountCents)
	}
}

func TestFinishTripPaymentFailureNotBlocking(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	pc := &fakePayClient{createPayment: func(_ context.Context, _ *pay.CreatePaymentRequest) (*pay.CreatePaymentResponse, error) {
		return nil, errors.New("paysvc down")
	}}
	l := NewFinishTripLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PayClient:       pc,
	})

	resp, err := l.FinishTrip(&proto.FinishTripRequest{
		OrderId:          int64(order.Id),
		DriverId:         2002,
		ActualPriceCents: 5200,
	})
	if err != nil {
		t.Fatalf("FinishTrip() error = %v, want success despite payment failure", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_WAIT_PAY {
		t.Fatalf("FinishTrip() status = %v, want WAIT_PAY", resp.Status)
	}
	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 4 {
		t.Fatalf("finished orderclient status = %d, want 4", fresh.Status)
	}
}

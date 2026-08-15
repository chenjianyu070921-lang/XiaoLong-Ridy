package logic

import (
	"context"
	"errors"
	"testing"

	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"google.golang.org/grpc"
)

type fakeDispatchClient struct {
	dispatchErr error
	called      bool
}

func (f *fakeDispatchClient) DispatchOrder(_ context.Context, in *dispatch.DispatchOrderRequest, _ ...grpc.CallOption) (*dispatch.DispatchOrderResponse, error) {
	f.called = true
	if f.dispatchErr != nil {
		return nil, f.dispatchErr
	}
	return &dispatch.DispatchOrderResponse{OrderId: in.OrderId}, nil
}

func (f *fakeDispatchClient) ListDispatchRecords(_ context.Context, _ *dispatch.ListDispatchRecordsRequest, _ ...grpc.CallOption) (*dispatch.ListDispatchRecordsResponse, error) {
	return &dispatch.ListDispatchRecordsResponse{}, nil
}

func validCreateOrderRequest() *proto.CreateOrderRequest {
	return &proto.CreateOrderRequest{
		UserId:              1001,
		CarType:             1,
		FromAddress:         "北京市朝阳区建国路88号",
		FromLongitude:       116.47319,
		FromLatitude:        39.9096,
		ToAddress:           "北京市海淀区中关村大街27号",
		ToLongitude:         116.31683,
		ToLatitude:          39.98472,
		EstimatedDistanceM:  12000,
		EstimatedDurationS:  1800,
		EstimatedPriceCents: 3600,
	}
}

func TestCreateOrderSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.CreateOrder(validCreateOrderRequest())
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if resp.OrderId <= 0 {
		t.Fatal("CreateOrder() returned empty order id")
	}
	if resp.OrderNo == "" {
		t.Fatal("CreateOrder() returned empty order no")
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT {
		t.Fatalf("CreateOrder() status = %v, want WAIT_ACCEPT", resp.Status)
	}
	if resp.EstimatedPriceCents != 3600 {
		t.Fatalf("CreateOrder() estimated price = %d, want 3600", resp.EstimatedPriceCents)
	}
	if resp.CreatedAt <= 0 {
		t.Fatal("CreateOrder() returned invalid created_at")
	}

	order, err := repo.GetByID(context.Background(), uint64(resp.OrderId))
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if order.UserId != 1001 || order.Status != 1 {
		t.Fatalf("stored order = %+v", order)
	}
	if order.EstimatedPrice != 36 {
		t.Fatalf("stored estimated price = %v, want 36", order.EstimatedPrice)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 1 || logs[0].ToStatus != 1 || logs[0].OperatorType != "user" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestCreateOrderRejectsInvalidParams(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	cases := []struct {
		name   string
		mutate func(*proto.CreateOrderRequest)
	}{
		{"zero user", func(r *proto.CreateOrderRequest) { r.UserId = 0 }},
		{"car type 0", func(r *proto.CreateOrderRequest) { r.CarType = 0 }},
		{"car type 4", func(r *proto.CreateOrderRequest) { r.CarType = 4 }},
		{"empty from", func(r *proto.CreateOrderRequest) { r.FromAddress = "  " }},
		{"empty to", func(r *proto.CreateOrderRequest) { r.ToAddress = "" }},
		{"bad from lng", func(r *proto.CreateOrderRequest) { r.FromLongitude = 181 }},
		{"bad from lat", func(r *proto.CreateOrderRequest) { r.FromLatitude = -91 }},
		{"bad to lng", func(r *proto.CreateOrderRequest) { r.ToLongitude = 200 }},
		{"bad to lat", func(r *proto.CreateOrderRequest) { r.ToLatitude = 90.1 }},
		{"negative distance", func(r *proto.CreateOrderRequest) { r.EstimatedDistanceM = -1 }},
		{"negative duration", func(r *proto.CreateOrderRequest) { r.EstimatedDurationS = -1 }},
		{"negative price", func(r *proto.CreateOrderRequest) { r.EstimatedPriceCents = -1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validCreateOrderRequest()
			tc.mutate(req)
			if _, err := l.CreateOrder(req); err != ErrInvalidOrderParams {
				t.Fatalf("CreateOrder() error = %v, want %v", err, ErrInvalidOrderParams)
			}
		})
	}
}

func TestCreateOrderTriggersDispatchAndIgnoresDispatchError(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	dispatchClient := &fakeDispatchClient{dispatchErr: errors.New("dispatch down")}
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		DispatchClient:  dispatchClient,
	})

	resp, err := l.CreateOrder(validCreateOrderRequest())
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if resp.OrderId <= 0 {
		t.Fatal("CreateOrder() returned empty order id")
	}
	if !dispatchClient.called {
		t.Fatal("CreateOrder() did not trigger dispatch")
	}
}

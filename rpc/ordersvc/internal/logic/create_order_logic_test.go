package logic

import (
	"context"
	"errors"
	"testing"

	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

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

type fakePriceClient struct {
	price.Price
	estimatePrice func(ctx context.Context, in *price.EstimatePriceRequest) (*price.EstimatePriceResponse, error)
	gotCityCode   string
}

// TestCreateOrder_RecordsUserBlacklistHit 验证下单用户命中黑名单时会写入下单场景审计记录。
func TestCreateOrder_RecordsUserBlacklistHit(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	risk := repository.NewMemoryRiskBlacklistRepository()
	risk.SetActive("user", 1001, repository.BlacklistEntry{ID: 88, Reason: "历史风险命中"})
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{
		OrderRepository:         repo,
		RiskBlacklistRepository: risk,
	})

	if _, err := l.CreateOrder(validCreateOrderRequest()); err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if len(risk.Hits) != 1 || risk.Hits[0].Scene != "order" || risk.Hits[0].BlacklistID != 88 || risk.Hits[0].TargetID != 1001 {
		t.Fatalf("risk hit records = %+v, want one order record", risk.Hits)
	}
}

func (f *fakePriceClient) EstimatePrice(_ context.Context, in *price.EstimatePriceRequest, _ ...grpc.CallOption) (*price.EstimatePriceResponse, error) {
	f.gotCityCode = in.CityCode
	if f.estimatePrice != nil {
		return f.estimatePrice(nil, in)
	}
	return &price.EstimatePriceResponse{TotalCents: 4500}, nil
}

func (f *fakePriceClient) CalculateDiscount(_ context.Context, _ *price.CalculateDiscountRequest, _ ...grpc.CallOption) (*price.CalculateDiscountResponse, error) {
	return nil, nil
}

func TestCreateOrderUsesPriceSnapshot(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	pc := &fakePriceClient{}
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PriceClient:     pc,
	})

	req := validCreateOrderRequest()
	req.CityCode = "310000"
	resp, err := l.CreateOrder(req)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if resp.EstimatedPriceCents != 4500 {
		t.Fatalf("EstimatedPriceCents = %d, want 4500", resp.EstimatedPriceCents)
	}
	if pc.gotCityCode != "310000" {
		t.Fatalf("price client city code = %q, want 310000", pc.gotCityCode)
	}
	order, err := repo.GetByID(context.Background(), uint64(resp.OrderId))
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if order.EstimatedPrice != 45 {
		t.Fatalf("stored estimated price = %v, want 45", order.EstimatedPrice)
	}
}

func TestCreateOrderFallbackOnPriceError(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	pc := &fakePriceClient{estimatePrice: func(_ context.Context, _ *price.EstimatePriceRequest) (*price.EstimatePriceResponse, error) {
		return nil, errors.New("pricesvc down")
	}}
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PriceClient:     pc,
	})

	resp, err := l.CreateOrder(validCreateOrderRequest())
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if resp.EstimatedPriceCents != 3600 {
		t.Fatalf("EstimatedPriceCents = %d, want fallback 3600", resp.EstimatedPriceCents)
	}
}

func TestCreateOrderPriceSnapshotDefaultCity(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	pc := &fakePriceClient{}
	l := NewCreateOrderLogic(context.Background(), &svc.ServiceContext{
		OrderRepository: repo,
		PriceClient:     pc,
	})

	// 不传 city_code，应兜底默认城市 110000
	if _, err := l.CreateOrder(validCreateOrderRequest()); err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if pc.gotCityCode != "110000" {
		t.Fatalf("price client city code = %q, want default 110000", pc.gotCityCode)
	}
}

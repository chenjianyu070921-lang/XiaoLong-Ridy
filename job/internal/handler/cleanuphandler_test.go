package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"XiaoLong-Ridy/job/internal/config"
	"XiaoLong-Ridy/job/internal/svc"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"google.golang.org/grpc"
)

// fakeOrderClient 只实现超时取消任务用到的两个方法，其余返回空。
type fakeOrderClient struct {
	pages     [][]*order.OrderSummary
	pageIndex int
	cancelErr error
	cancelled []int64
	timeouts  []int32
}

func (f *fakeOrderClient) ListTimeoutOrders(_ context.Context, in *order.ListTimeoutOrdersRequest, _ ...grpc.CallOption) (*order.ListTimeoutOrdersResponse, error) {
	f.timeouts = append(f.timeouts, in.TimeoutSeconds)
	if f.pageIndex >= len(f.pages) {
		return &order.ListTimeoutOrdersResponse{List: []*order.OrderSummary{}, Page: in.Page, PageSize: in.PageSize}, nil
	}
	items := f.pages[f.pageIndex]
	f.pageIndex++
	return &order.ListTimeoutOrdersResponse{List: items, Total: int64(len(items)), Page: in.Page, PageSize: in.PageSize}, nil
}

func (f *fakeOrderClient) TimeoutCancel(_ context.Context, in *order.TimeoutCancelRequest, _ ...grpc.CallOption) (*order.TimeoutCancelResponse, error) {
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	f.cancelled = append(f.cancelled, in.OrderId)
	return &order.TimeoutCancelResponse{OrderId: in.OrderId}, nil
}

func (f *fakeOrderClient) CreateOrder(context.Context, *order.CreateOrderRequest, ...grpc.CallOption) (*order.CreateOrderResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) CancelOrder(context.Context, *order.CancelOrderRequest, ...grpc.CallOption) (*order.CancelOrderResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) GetOrder(context.Context, *order.GetOrderRequest, ...grpc.CallOption) (*order.GetOrderResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) ListOrders(context.Context, *order.ListOrdersRequest, ...grpc.CallOption) (*order.ListOrdersResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) AcceptOrder(context.Context, *order.AcceptOrderRequest, ...grpc.CallOption) (*order.AcceptOrderResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) ConfirmArrive(context.Context, *order.ConfirmArriveRequest, ...grpc.CallOption) (*order.ConfirmArriveResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) StartTrip(context.Context, *order.StartTripRequest, ...grpc.CallOption) (*order.StartTripResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) FinishTrip(context.Context, *order.FinishTripRequest, ...grpc.CallOption) (*order.FinishTripResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) ConfirmPaid(context.Context, *order.ConfirmPaidRequest, ...grpc.CallOption) (*order.ConfirmPaidResponse, error) {
	return nil, nil
}
func (f *fakeOrderClient) ListOrderStatusLogs(context.Context, *order.ListOrderStatusLogsRequest, ...grpc.CallOption) (*order.ListOrderStatusLogsResponse, error) {
	return nil, nil
}

func newTestHandler(client order.Order) *CleanupHandler {
	return NewCleanupHandler(&svc.ServiceContext{
		Config:      config.Config{TimeoutSeconds: 0},
		OrderClient: client,
	})
}

func makeSummary(orderID int64) *order.OrderSummary {
	return &order.OrderSummary{OrderId: orderID, OrderNo: fmt.Sprintf("NO%d", orderID)}
}

func TestTimeoutCancelOrders_MultiPage(t *testing.T) {
	page1 := make([]*order.OrderSummary, 0, 50)
	for i := 0; i < 50; i++ {
		page1 = append(page1, makeSummary(int64(i+1)))
	}
	fake := &fakeOrderClient{
		pages: [][]*order.OrderSummary{
			page1,
			{makeSummary(51), makeSummary(52)},
		},
	}
	h := newTestHandler(fake)

	if err := h.TimeoutCancelOrders(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.cancelled) != 52 {
		t.Fatalf("expect 52 cancelled, got %d", len(fake.cancelled))
	}
	if len(fake.timeouts) != 2 {
		t.Fatalf("expect 2 pagination calls, got %d", len(fake.timeouts))
	}
	// 配置为 0 时应使用默认 300 秒
	for i, v := range fake.timeouts {
		if v != 300 {
			t.Fatalf("page %d: expect timeout_seconds=300, got %d", i, v)
		}
	}
}

func TestTimeoutCancelOrders_ConfiguredTimeout(t *testing.T) {
	fake := &fakeOrderClient{pages: [][]*order.OrderSummary{{makeSummary(1)}}}
	h := NewCleanupHandler(&svc.ServiceContext{
		Config:      config.Config{TimeoutSeconds: 600},
		OrderClient: fake,
	})
	if err := h.TimeoutCancelOrders(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.timeouts[0] != 600 {
		t.Fatalf("expect timeout_seconds=600, got %d", fake.timeouts[0])
	}
}

func TestTimeoutCancelOrders_CancelErrorNotBlocking(t *testing.T) {
	fake := &fakeOrderClient{
		pages:     [][]*order.OrderSummary{{makeSummary(1), makeSummary(2)}},
		cancelErr: errors.New("status not cancelable"),
	}
	h := newTestHandler(fake)

	if err := h.TimeoutCancelOrders(); err != nil {
		t.Fatalf("cancel failure should not block the task: %v", err)
	}
	if len(fake.cancelled) != 0 {
		t.Fatalf("expect 0 cancelled, got %d", len(fake.cancelled))
	}
}

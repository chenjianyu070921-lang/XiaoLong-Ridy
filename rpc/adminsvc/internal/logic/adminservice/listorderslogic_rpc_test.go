package adminservicelogic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	ordersvc "XiaoLong-Ridy/rpc/ordersvc/proto"

	"google.golang.org/grpc"
)

// fakeOrdersClient 提供订单列表 RPC 的最小测试替身。
type fakeOrdersClient struct {
	ordersvc.OrderClient
	request        *ordersvc.ListOrdersRequest
	detailRequest  *ordersvc.GetOrderRequest
	statusLogError error
}

// ListOrders 记录后台转发给 ordersvc 的查询条件。
func (f *fakeOrdersClient) ListOrders(_ context.Context, in *ordersvc.ListOrdersRequest, _ ...grpc.CallOption) (*ordersvc.ListOrdersResponse, error) {
	f.request = in
	return &ordersvc.ListOrdersResponse{
		List: []*ordersvc.OrderSummary{{
			OrderId: 1001, OrderNo: "RO1001", FromAddress: "起点", ToAddress: "终点",
			UserId: 88, DriverId: 66, CarType: 2, FromLongitude: 116.123456, FromLatitude: 39.123456,
			ToLongitude: 116.654321, ToLatitude: 39.654321, EstimatedDistanceM: 12345, EstimatedDurationS: 1800,
			Status: ordersvc.OrderStatus_ORDER_STATUS_WAIT_ACCEPT, EstimatedPriceCents: 3500, CreatedAt: 1724150400, UpdatedAt: 1724150500,
		}},
		Total: 1, Page: 1, PageSize: 20,
	}, nil
}

// GetOrder 记录后台订单详情对 ordersvc 的主订单查询。
func (f *fakeOrdersClient) GetOrder(_ context.Context, in *ordersvc.GetOrderRequest, _ ...grpc.CallOption) (*ordersvc.GetOrderResponse, error) {
	f.detailRequest = in
	return &ordersvc.GetOrderResponse{
		OrderId: 1001, OrderNo: "RO1001", UserId: 88, DriverId: 66, CarType: 2,
		FromAddress: "起点", FromLongitude: 116.123456, FromLatitude: 39.123456,
		ToAddress: "终点", ToLongitude: 116.654321, ToLatitude: 39.654321,
		EstimatedDistanceM: 12345, EstimatedDurationS: 1800, EstimatedPriceCents: 3500,
		Status: ordersvc.OrderStatus_ORDER_STATUS_WAIT_ACCEPT, CreatedAt: 1724150400, UpdatedAt: 1724150500,
	}, nil
}

// ListOrderStatusLogs 返回测试配置的状态流水结果。
func (f *fakeOrdersClient) ListOrderStatusLogs(_ context.Context, in *ordersvc.ListOrderStatusLogsRequest, _ ...grpc.CallOption) (*ordersvc.ListOrderStatusLogsResponse, error) {
	if f.statusLogError != nil {
		return nil, f.statusLogError
	}
	return &ordersvc.ListOrderStatusLogsResponse{List: []*ordersvc.OrderStatusLog{{Id: 1, OrderId: in.GetOrderId(), ToStatus: 1, OperatorType: "user", CreatedAt: 1724150400}}}, nil
}

// TestListOrders_UsesOrdersRPCWhenFiltersAreSupported 验证可表达筛选条件优先走真实 ordersvc。
func TestListOrders_UsesOrdersRPCWhenFiltersAreSupported(t *testing.T) {
	client := &fakeOrdersClient{}
	logic := NewListOrdersLogic(context.Background(), &svc.ServiceContext{OrdersSvc: client})
	resp, err := logic.ListOrders(&adminsvc.OrderListRequest{Page: 1, PageSize: 20, UserId: 88, Status: 1})
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if client.request == nil || client.request.GetUserId() != 88 || client.request.GetStatus() != ordersvc.OrderStatus_ORDER_STATUS_WAIT_ACCEPT {
		t.Fatalf("ordersvc request = %+v", client.request)
	}
	if resp.GetTotal() != 1 || len(resp.GetList()) != 1 || resp.GetList()[0].GetOrderNo() != "RO1001" {
		t.Fatalf("response = %+v", resp)
	}
	got := resp.GetList()[0]
	if got.GetUserId() != 88 || got.GetDriverId() != 66 || got.GetCarType() != 2 ||
		got.GetFromLongitude() != "116.123456" || got.GetEstimatedDistanceM() != 12345 || got.GetUpdatedAt() == "" {
		t.Fatalf("order fields = %+v", got)
	}
}

// TestOrderListCanUseOrdersRPC_DisablesUnsupportedFilters 验证关键字和时间筛选不会被错误丢弃。
func TestOrderListCanUseOrdersRPC_DisablesUnsupportedFilters(t *testing.T) {
	if orderListCanUseOrdersRPC(&adminsvc.OrderListRequest{Keyword: "RO1001"}) {
		t.Fatal("keyword filter should use compatibility query")
	}
	if orderListCanUseOrdersRPC(&adminsvc.OrderListRequest{StartTime: "2026-08-25 00:00:00"}) {
		t.Fatal("time filter should use compatibility query")
	}
}

// TestGetOrder_UsesOrdersRPCMainData 验证订单详情主信息以 ordersvc.GetOrder 返回为准。
func TestGetOrder_UsesOrdersRPCMainData(t *testing.T) {
	client := &fakeOrdersClient{}
	logic := NewGetOrderLogic(context.Background(), &svc.ServiceContext{OrdersSvc: client})
	resp, err := logic.GetOrder(&adminsvc.OrderDetailRequest{Id: 1001})
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if client.detailRequest == nil || client.detailRequest.GetOrderId() != 1001 {
		t.Fatalf("ordersvc detail request = %+v", client.detailRequest)
	}
	if resp.GetOrder().GetUserId() != 88 || resp.GetOrder().GetDriverId() != 66 || resp.GetOrder().GetFromLongitude() != "116.123456" {
		t.Fatalf("GetOrder() order = %+v", resp.GetOrder())
	}
	if len(resp.GetStatusLogs()) != 1 {
		t.Fatalf("status logs = %+v", resp.GetStatusLogs())
	}
	if len(resp.GetDegraded()) != 3 {
		t.Fatalf("degraded = %+v, want local mysql dependent modules", resp.GetDegraded())
	}
}

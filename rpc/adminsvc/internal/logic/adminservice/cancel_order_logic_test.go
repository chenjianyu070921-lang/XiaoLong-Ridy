package adminservicelogic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"google.golang.org/grpc"
)

// fakeOrderClient 只实现本测试关注的 CancelOrder 调用，其余方法由嵌入接口占位。
type fakeOrderClient struct {
	orderproto.OrderClient
	got *orderproto.CancelOrderRequest
}

// CancelOrder 记录 adminsvc 转发给 ordersvc 的请求参数。
func (f *fakeOrderClient) CancelOrder(ctx context.Context, in *orderproto.CancelOrderRequest, opts ...grpc.CallOption) (*orderproto.CancelOrderResponse, error) {
	f.got = in
	return &orderproto.CancelOrderResponse{OrderId: in.GetOrderId(), Status: orderproto.OrderStatus_ORDER_STATUS_CANCELLED}, nil
}

// TestCancelOrderLogic_CallsOrdersvc 验证后台取消订单逻辑会把请求转发给 ordersvc。
func TestCancelOrderLogic_CallsOrdersvc(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrdersSvc: client})
	resp, err := logic.CancelOrder(&adminsvc.AdminCancelOrderRequest{
		OrderId: 1001,
		Reason:  "运营人工取消",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp == nil || resp.Message != "ok" {
		t.Fatalf("CancelOrder() response = %#v, want ok", resp)
	}
	if client.got == nil {
		t.Fatal("CancelOrder() did not call ordersvc")
	}
	if client.got.GetOrderId() != 1001 || client.got.GetOperatorId() != 9001 || client.got.GetOperatorType() != "admin" {
		t.Fatalf("ordersvc request = %#v, want order/admin operator fields", client.got)
	}
	if client.got.GetReason() != "运营人工取消" {
		t.Fatalf("ordersvc reason = %q, want %q", client.got.GetReason(), "运营人工取消")
	}
}

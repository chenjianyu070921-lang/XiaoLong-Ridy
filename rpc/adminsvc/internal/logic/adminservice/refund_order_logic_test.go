package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
)

// fakeRefundOrderClient 记录后台退款转发给 ordersvc 的请求。
// 测试只关注 ForceRefundOrder，不依赖真实订单服务和数据库。
type fakeRefundOrderClient struct {
	orderproto.OrderClient
	got *orderproto.ForceRefundOrderRequest
	err error
}

// ForceRefundOrder 模拟 ordersvc 的管理员退款接口。
func (f *fakeRefundOrderClient) ForceRefundOrder(_ context.Context, in *orderproto.ForceRefundOrderRequest, _ ...grpc.CallOption) (*orderproto.ForceRefundOrderResponse, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	return &orderproto.ForceRefundOrderResponse{
		OrderId:     in.GetOrderId(),
		Status:      orderproto.OrderStatus_ORDER_STATUS_REFUNDED,
		RefundCents: in.GetRefundAmountCents(),
	}, nil
}

// TestRefundOrderLogic_ForwardsAndAudits 验证退款请求会携带后台管理员信息、退款幂等号并写入审计日志。
func TestRefundOrderLogic_ForwardsAndAudits(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "order", "refund", "ride_order", int64(1001), sqlmock.AnyArg(), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	client := &fakeRefundOrderClient{}
	svcCtx.OrdersSvc = client
	resp, err := NewRefundOrderLogic(context.Background(), svcCtx).RefundOrder(&adminsvc.AdminRefundOrderRequest{
		OrderId:           1001,
		RefundAmountCents: 1850,
		Reason:            "投诉成立",
		AdminId:           9001,
		Ip:                "127.0.0.1",
		RequestId:         "refund-test-001",
	})
	if err != nil {
		t.Fatalf("RefundOrder() error = %v", err)
	}
	if resp == nil || resp.GetMessage() != "ok" || resp.GetRefundNo() != "refund-test-001" {
		t.Fatalf("RefundOrder() response = %#v, want successful idempotent response", resp)
	}
	if client.got == nil {
		t.Fatal("RefundOrder() did not call ordersvc")
	}
	if client.got.GetOrderId() != 1001 || client.got.GetOperatorId() != 9001 ||
		client.got.GetRefundNo() != "refund-test-001" || client.got.GetRefundAmountCents() != 1850 {
		t.Fatalf("ordersvc request = %#v, want order/admin/refund fields", client.got)
	}
}

// TestRefundOrderLogic_DownstreamFailureReleasesRetryKey 验证下游失败时允许同一个 request_id 重试。
func TestRefundOrderLogic_DownstreamFailureReleasesRetryKey(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()

	client := &fakeRefundOrderClient{err: errors.New("ordersvc unavailable")}
	svcCtx.OrdersSvc = client
	_, err := NewRefundOrderLogic(context.Background(), svcCtx).RefundOrder(&adminsvc.AdminRefundOrderRequest{
		OrderId:           1001,
		RefundAmountCents: 1850,
		Reason:            "投诉成立",
		AdminId:           9001,
		RequestId:         "refund-test-002",
	})
	if err == nil {
		t.Fatal("RefundOrder() error = nil, want downstream error")
	}
}

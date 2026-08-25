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

// fakeOrderClient 只实现本测试关注的 CancelOrder 调用，其余方法由嵌入接口占位。
type fakeOrderClient struct {
	orderproto.OrderClient
	got *orderproto.CancelOrderRequest
	err error
}

// CancelOrder 记录 adminsvc 转发给 ordersvc 的请求参数。
func (f *fakeOrderClient) CancelOrder(ctx context.Context, in *orderproto.CancelOrderRequest, opts ...grpc.CallOption) (*orderproto.CancelOrderResponse, error) {
	f.got = in
	if f.err != nil {
		return nil, f.err
	}
	return &orderproto.CancelOrderResponse{OrderId: in.GetOrderId(), Status: orderproto.OrderStatus_ORDER_STATUS_CANCELLED}, nil
}

// TestCancelOrderLogic_CallsOrdersvc 验证后台取消订单逻辑会把请求转发给 ordersvc。
func TestCancelOrderLogic_CallsOrdersvc(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "order", "cancel", "ride_order", int64(1001), "后台取消订单：运营人工取消", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	client := &fakeOrderClient{}
	svcCtx.OrdersSvc = client
	logic := NewCancelOrderLogic(context.Background(), svcCtx)
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
		t.Fatalf("ordersvc request = %#v, want orderclient/admin operator fields", client.got)
	}
	if client.got.GetReason() != "运营人工取消" {
		t.Fatalf("ordersvc reason = %q, want %q", client.got.GetReason(), "运营人工取消")
	}
}

// TestCancelOrderLogic_AuditFailureCreatesOutbox 验证审计日志失败时会创建补偿任务并保持业务成功语义。
func TestCancelOrderLogic_AuditFailureCreatesOutbox(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "order", "cancel", "ride_order", int64(1001), int64(9001),
			"后台取消订单：运营人工取消", "127.0.0.1", "operation log write failed").
		WillReturnResult(sqlmock.NewResult(1, 1))

	svcCtx.OrdersSvc = &fakeOrderClient{}
	resp, err := NewCancelOrderLogic(context.Background(), svcCtx).CancelOrder(&adminsvc.AdminCancelOrderRequest{
		OrderId: 1001,
		Reason:  "运营人工取消",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil || resp == nil || resp.GetMessage() != "ok" {
		t.Fatalf("CancelOrder() = %#v, %v; want successful compensated response", resp, err)
	}
}

// TestCancelOrderLogic_AuditAndOutboxFailureReturnsError 验证审计和补偿均失败时不会返回业务成功。
func TestCancelOrderLogic_AuditAndOutboxFailureReturnsError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WillReturnError(errors.New("outbox write failed"))

	svcCtx.OrdersSvc = &fakeOrderClient{}
	resp, err := NewCancelOrderLogic(context.Background(), svcCtx).CancelOrder(&adminsvc.AdminCancelOrderRequest{
		OrderId: 1001,
		Reason:  "运营人工取消",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err == nil {
		t.Fatal("CancelOrder() error = nil, want audit and outbox error")
	}
	if resp != nil {
		t.Fatalf("CancelOrder() response = %#v, want nil", resp)
	}
}

package consumer

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/mq-consumer/order-event-consumer/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"
	pay "XiaoLong-Ridy/rpc/paysvc/pay"

	"google.golang.org/grpc"
)

// mockDispatch 实现 dispatch.Dispatch，记录调用。
type mockDispatch struct {
	dispatchCalls int
	lastReq       *dispatch.DispatchOrderRequest
}

func (m *mockDispatch) DispatchOrder(ctx context.Context, in *dispatch.DispatchOrderRequest, opts ...grpc.CallOption) (*dispatch.DispatchOrderResponse, error) {
	m.dispatchCalls++
	m.lastReq = in
	return &dispatch.DispatchOrderResponse{OrderId: in.OrderId}, nil
}
func (m *mockDispatch) ListDispatchRecords(ctx context.Context, in *dispatch.ListDispatchRecordsRequest, opts ...grpc.CallOption) (*dispatch.ListDispatchRecordsResponse, error) {
	return &dispatch.ListDispatchRecordsResponse{}, nil
}
func (m *mockDispatch) RejectDispatch(ctx context.Context, in *dispatch.RejectDispatchRequest, opts ...grpc.CallOption) (*dispatch.RejectDispatchResponse, error) {
	return &dispatch.RejectDispatchResponse{}, nil
}
func (m *mockDispatch) CancelDispatch(ctx context.Context, in *dispatch.CancelDispatchRequest, opts ...grpc.CallOption) (*dispatch.CancelDispatchResponse, error) {
	return &dispatch.CancelDispatchResponse{}, nil
}
func (m *mockDispatch) ListTimeoutPendingOrders(ctx context.Context, in *dispatch.ListTimeoutPendingOrdersRequest, opts ...grpc.CallOption) (*dispatch.ListTimeoutPendingOrdersResponse, error) {
	return &dispatch.ListTimeoutPendingOrdersResponse{}, nil
}

// mockOrder 实现 order.Order，记录调用。
type mockOrder struct {
	paidCalls int
	lastReq   *order.ConfirmPaidRequest
}

func (m *mockOrder) CreateOrder(ctx context.Context, in *order.CreateOrderRequest, opts ...grpc.CallOption) (*order.CreateOrderResponse, error) {
	return &order.CreateOrderResponse{}, nil
}
func (m *mockOrder) CancelOrder(ctx context.Context, in *order.CancelOrderRequest, opts ...grpc.CallOption) (*order.CancelOrderResponse, error) {
	return &order.CancelOrderResponse{}, nil
}
func (m *mockOrder) GetOrder(ctx context.Context, in *order.GetOrderRequest, opts ...grpc.CallOption) (*order.GetOrderResponse, error) {
	return &order.GetOrderResponse{}, nil
}
func (m *mockOrder) ListOrders(ctx context.Context, in *order.ListOrdersRequest, opts ...grpc.CallOption) (*order.ListOrdersResponse, error) {
	return &order.ListOrdersResponse{}, nil
}
func (m *mockOrder) AcceptOrder(ctx context.Context, in *order.AcceptOrderRequest, opts ...grpc.CallOption) (*order.AcceptOrderResponse, error) {
	return &order.AcceptOrderResponse{}, nil
}
func (m *mockOrder) ConfirmArrive(ctx context.Context, in *order.ConfirmArriveRequest, opts ...grpc.CallOption) (*order.ConfirmArriveResponse, error) {
	return &order.ConfirmArriveResponse{}, nil
}
func (m *mockOrder) StartTrip(ctx context.Context, in *order.StartTripRequest, opts ...grpc.CallOption) (*order.StartTripResponse, error) {
	return &order.StartTripResponse{}, nil
}
func (m *mockOrder) FinishTrip(ctx context.Context, in *order.FinishTripRequest, opts ...grpc.CallOption) (*order.FinishTripResponse, error) {
	return &order.FinishTripResponse{}, nil
}
func (m *mockOrder) ConfirmPaid(ctx context.Context, in *order.ConfirmPaidRequest, opts ...grpc.CallOption) (*order.ConfirmPaidResponse, error) {
	m.paidCalls++
	m.lastReq = in
	return &order.ConfirmPaidResponse{}, nil
}
func (m *mockOrder) TimeoutCancel(ctx context.Context, in *order.TimeoutCancelRequest, opts ...grpc.CallOption) (*order.TimeoutCancelResponse, error) {
	return &order.TimeoutCancelResponse{}, nil
}
func (m *mockOrder) ListTimeoutOrders(ctx context.Context, in *order.ListTimeoutOrdersRequest, opts ...grpc.CallOption) (*order.ListTimeoutOrdersResponse, error) {
	return &order.ListTimeoutOrdersResponse{}, nil
}
func (m *mockOrder) ListOrderStatusLogs(ctx context.Context, in *order.ListOrderStatusLogsRequest, opts ...grpc.CallOption) (*order.ListOrderStatusLogsResponse, error) {
	return &order.ListOrderStatusLogsResponse{}, nil
}
func (m *mockOrder) RefundOrder(ctx context.Context, in *order.RefundOrderRequest, opts ...grpc.CallOption) (*order.RefundOrderResponse, error) {
	return &order.RefundOrderResponse{}, nil
}
func (m *mockOrder) RedispatchOrder(ctx context.Context, in *order.RedispatchOrderRequest, opts ...grpc.CallOption) (*order.RedispatchOrderResponse, error) {
	return &order.RedispatchOrderResponse{}, nil
}
func (m *mockOrder) ForceRefundOrder(ctx context.Context, in *order.ForceRefundOrderRequest, opts ...grpc.CallOption) (*order.ForceRefundOrderResponse, error) {
	return &order.ForceRefundOrderResponse{}, nil
}

// mockPay 实现 pay.Pay，记录退款链路调用。
type mockPay struct {
	getPaymentCalls    int
	refundPaymentCalls int
	lastRefund         *pay.RefundPaymentRequest
	getPayment         *pay.GetPaymentResponse
	getPaymentErr      error
	refundErr          error
}

func (m *mockPay) CreatePayment(context.Context, *pay.CreatePaymentRequest, ...grpc.CallOption) (*pay.CreatePaymentResponse, error) {
	return &pay.CreatePaymentResponse{}, nil
}
func (m *mockPay) NotifyPayment(context.Context, *pay.NotifyPaymentRequest, ...grpc.CallOption) (*pay.NotifyPaymentResponse, error) {
	return &pay.NotifyPaymentResponse{}, nil
}
func (m *mockPay) GetPayment(_ context.Context, in *pay.GetPaymentRequest, _ ...grpc.CallOption) (*pay.GetPaymentResponse, error) {
	m.getPaymentCalls++
	return m.getPayment, m.getPaymentErr
}
func (m *mockPay) RefundPayment(_ context.Context, in *pay.RefundPaymentRequest, _ ...grpc.CallOption) (*pay.RefundPaymentResponse, error) {
	m.refundPaymentCalls++
	m.lastRefund = in
	if m.refundErr != nil {
		return nil, m.refundErr
	}
	return &pay.RefundPaymentResponse{Success: true, RefundNo: in.RefundNo}, nil
}
func (m *mockPay) SettleOrder(context.Context, *pay.SettleOrderRequest, ...grpc.CallOption) (*pay.SettleOrderResponse, error) {
	return &pay.SettleOrderResponse{}, nil
}

// newTestConsumer 构造测试用 OrderConsumer：Redis 用 miniredis 模拟，RPC 用 mock。
func newTestConsumer(t *testing.T) (*OrderConsumer, *mockDispatch, *mockOrder, *mockPay, *miniredis.Miniredis) {
	t.Helper()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis start error: %v", err)
	}
	md := &mockDispatch{}
	mo := &mockOrder{}
	mp := &mockPay{}
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() {
		rdb.Close()
		s.Close()
	})
	c := &OrderConsumer{svcCtx: &svc.ServiceContext{
		Redis:          rdb,
		DispatchClient: md,
		OrderClient:    mo,
		PayClient:      mp,
	}}
	return c, md, mo, mp, s
}

// TestDispatchHandler_OrderCreated order.created 事件应触发派单 RPC。
func TestDispatchHandler_OrderCreated(t *testing.T) {
	c, md, _, _, _ := newTestConsumer(t)
	payload := []byte(`{"order_id":101,"order_no":"XL2026","from_longitude":116.4,"from_latitude":39.9,"car_type":1,"city_code":"110000"}`)
	if err := c.dispatchHandler(context.Background(), constants.TopicOrderCreated, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md.dispatchCalls != 1 {
		t.Fatalf("DispatchOrder calls = %d, want 1", md.dispatchCalls)
	}
	if md.lastReq == nil || md.lastReq.OrderId != 101 || md.lastReq.CityCode != "110000" {
		t.Fatalf("DispatchOrder request = %+v, want order 101 city 110000", md.lastReq)
	}
}

// TestDispatchHandler_OrderPaid order.paid 事件应触发确认支付 RPC。
func TestDispatchHandler_OrderPaid(t *testing.T) {
	c, _, mo, _, _ := newTestConsumer(t)
	payload := []byte(`{"order_id":202,"payment_no":"P2026","amount_cents":4500,"paid_at":1700000000}`)
	if err := c.dispatchHandler(context.Background(), constants.TopicOrderPaid, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mo.paidCalls != 1 {
		t.Fatalf("ConfirmPaid calls = %d, want 1", mo.paidCalls)
	}
	if mo.lastReq == nil || mo.lastReq.OrderId != 202 || mo.lastReq.AmountCents != 4500 {
		t.Fatalf("ConfirmPaid request = %+v, want order 202 amount 4500", mo.lastReq)
	}
}

// TestDispatchHandler_OrderRefunded 退款事件应先查询支付单号，再调用支付退款。
func TestDispatchHandler_OrderRefunded(t *testing.T) {
	c, _, _, mp, _ := newTestConsumer(t)
	mp.getPayment = &pay.GetPaymentResponse{OrderId: 202, PaymentNo: "PAY2026"}
	payload := []byte(`{"order_id":202,"order_no":"XL2026","refund_no":"RF2026","refund_cents":4500,"operator_id":9,"operator_type":"admin"}`)

	if err := c.dispatchHandler(context.Background(), constants.TopicOrderRefunded, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mp.getPaymentCalls != 1 || mp.refundPaymentCalls != 1 {
		t.Fatalf("pay calls = get:%d refund:%d, want 1/1", mp.getPaymentCalls, mp.refundPaymentCalls)
	}
	if mp.lastRefund == nil || mp.lastRefund.PaymentNo != "PAY2026" ||
		mp.lastRefund.RefundAmountCents != 4500 || mp.lastRefund.RefundNo != "RF2026" {
		t.Fatalf("refund request = %+v, want payment PAY2026 amount 4500 refund RF2026", mp.lastRefund)
	}
}

// TestDispatchHandler_OrderRefundedWithoutPayment 关联支付单不存在时必须返回错误，交由 DLQ 处理。
func TestDispatchHandler_OrderRefundedWithoutPayment(t *testing.T) {
	c, _, _, mp, _ := newTestConsumer(t)
	mp.getPayment = &pay.GetPaymentResponse{}
	payload := []byte(`{"order_id":202,"refund_no":"RF2026","refund_cents":4500}`)

	if err := c.dispatchHandler(context.Background(), constants.TopicOrderRefunded, payload); err == nil {
		t.Fatal("expected payment not found error")
	}
	if mp.refundPaymentCalls != 0 {
		t.Fatalf("RefundPayment calls = %d, want 0", mp.refundPaymentCalls)
	}
}

// TestDispatchHandler_DispatchNew dispatch.new 事件应把订单写入候选司机的待接单列表。
func TestDispatchHandler_DispatchNew(t *testing.T) {
	c, _, _, _, s := newTestConsumer(t)
	payload := []byte(`{"order_id":303,"driver_ids":[1,2],"from_longitude":116.4,"from_latitude":39.9,"car_type":1,"city_code":"110000","dispatched_at":1700000000}`)
	if err := c.dispatchHandler(context.Background(), constants.TopicDispatchNew, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, driverID := range []int64{1, 2} {
		if !s.Exists(availableListKey(driverID)) {
			t.Fatalf("available list for driver %d not found", driverID)
		}
		members, err := s.SMembers(availableListKey(driverID))
		if err != nil {
			t.Fatalf("SMembers error: %v", err)
		}
		if len(members) != 1 || members[0] != "303" {
			t.Fatalf("driver %d available members = %v, want [303]", driverID, members)
		}
	}
}

// TestDispatchHandler_UnknownTopic 未知 topic 应直接忽略，不报错。
func TestDispatchHandler_UnknownTopic(t *testing.T) {
	c, _, _, _, _ := newTestConsumer(t)
	if err := c.dispatchHandler(context.Background(), "unknown.topic", []byte(`{}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDispatchHandler_BadPayload 非法载荷应返回解析错误（交由消费者 DLQ 兜底）。
func TestDispatchHandler_BadPayload(t *testing.T) {
	c, _, _, _, _ := newTestConsumer(t)
	if err := c.dispatchHandler(context.Background(), constants.TopicOrderCreated, []byte(`not-json`)); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

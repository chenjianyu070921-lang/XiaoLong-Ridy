package svc

import (
	"context"
	"sync"

	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
)

// localPaymentCreator 抽象本地支付创建能力，避免乘客端依赖 paysvc 的内部实现细节。
type localPaymentCreator interface {
	CreatePayment(ctx context.Context, req *payproto.CreatePaymentRequest) (*payproto.CreatePaymentResponse, error)
}

// localPayClient 为显式 local 模式补齐支付查询能力，支付单只保存在当前进程内存中。
type localPayClient struct {
	creator localPaymentCreator
	mu      sync.Mutex
	byNo    map[string]*payproto.GetPaymentResponse
	byOrder map[int64]*payproto.GetPaymentResponse
}

// newLocalPayClient 创建乘客端本地支付适配器，保持 local 模式可自包含演示支付入口和查询入口。
func newLocalPayClient(creator localPaymentCreator) *localPayClient {
	return &localPayClient{
		creator: creator,
		byNo:    make(map[string]*payproto.GetPaymentResponse),
		byOrder: make(map[int64]*payproto.GetPaymentResponse),
	}
}

// CreatePayment 创建本地支付单并缓存查询快照。
func (c *localPayClient) CreatePayment(ctx context.Context, req *payproto.CreatePaymentRequest) (*payproto.CreatePaymentResponse, error) {
	resp, err := c.creator.CreatePayment(ctx, req)
	if err != nil {
		return nil, err
	}
	payment := &payproto.GetPaymentResponse{
		PaymentId:   resp.GetPaymentId(),
		PaymentNo:   resp.GetPaymentNo(),
		OrderId:     req.GetOrderId(),
		AmountCents: req.GetAmountCents(),
		Channel:     req.GetChannel().String(),
		// local 模式没有真实第三方回调，创建成功即视为支付成功，避免页面永久轮询待支付状态。
		Status:        2,
		TransactionId: resp.GetTransactionId(),
	}
	c.mu.Lock()
	c.byNo[payment.PaymentNo] = clonePayment(payment)
	c.byOrder[payment.OrderId] = clonePayment(payment)
	c.mu.Unlock()
	return resp, nil
}

// GetPayment 按支付单号或订单 ID 查询本地支付单快照。
func (c *localPayClient) GetPayment(_ context.Context, req *payproto.GetPaymentRequest) (*payproto.GetPaymentResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req.GetPaymentNo() != "" {
		if payment, ok := c.byNo[req.GetPaymentNo()]; ok {
			return clonePayment(payment), nil
		}
	}
	if req.GetOrderId() > 0 {
		if payment, ok := c.byOrder[req.GetOrderId()]; ok {
			return clonePayment(payment), nil
		}
	}
	return &payproto.GetPaymentResponse{}, nil
}

// clonePayment 复制支付查询结果，避免调用方修改内存缓存。
func clonePayment(payment *payproto.GetPaymentResponse) *payproto.GetPaymentResponse {
	if payment == nil {
		return nil
	}
	return &payproto.GetPaymentResponse{
		PaymentId:         payment.GetPaymentId(),
		PaymentNo:         payment.GetPaymentNo(),
		OrderId:           payment.GetOrderId(),
		AmountCents:       payment.GetAmountCents(),
		Channel:           payment.GetChannel(),
		Status:            payment.GetStatus(),
		TransactionId:     payment.GetTransactionId(),
		RefundAmountCents: payment.GetRefundAmountCents(),
	}
}

// memoryDispatchClient 是 local 模式下的派单查询占位实现，真实派单仍由 dispatchsvc 负责。
type memoryDispatchClient struct{}

// newMemoryDispatchClient 创建本地派单查询客户端。
func newMemoryDispatchClient() *memoryDispatchClient {
	return &memoryDispatchClient{}
}

// ListDispatchRecords 返回空派单记录，保证无 dispatchsvc 时乘客端状态查询仍可稳定响应。
func (c *memoryDispatchClient) ListDispatchRecords(_ context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	return &dispatchproto.ListDispatchRecordsResponse{
		List:     []*dispatchproto.DispatchRecord{},
		Total:    0,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

package pay

import (
	"context"
	"fmt"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/proto"
)

// LocalClient 是本地开发和测试使用的支付服务实现。
type LocalClient struct {
	mu     sync.Mutex
	nextID int64
}

// NewLocalClient 创建本地支付客户端，仅在 passenger 显式 local 模式下使用。
func NewLocalClient() *LocalClient {
	return &LocalClient{nextID: 1}
}

// CreatePayment 创建本地 mock 支付单，返回与 paysvc 一致的支付参数结构。
func (c *LocalClient) CreatePayment(_ context.Context, req *proto.CreatePaymentRequest) (*proto.CreatePaymentResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	paymentID := c.nextID
	c.nextID++
	paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), paymentID)
	return &proto.CreatePaymentResponse{
		PaymentId:     paymentID,
		PaymentNo:     paymentNo,
		TransactionId: fmt.Sprintf("LOCAL_TX_%06d", paymentID),
		PayParams:     fmt.Sprintf("local://pay/%s?amount=%d&channel=%d", paymentNo, req.GetAmountCents(), req.GetChannel()),
		Status:        1,
	}, nil
}

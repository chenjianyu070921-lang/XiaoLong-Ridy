package repository

import "context"

// MemoryCouponConsumer 是 ConfirmPaid 单元测试使用的优惠券核销替身。
type MemoryCouponConsumer struct {
	Calls   int
	UserID  uint64
	OrderID uint64
	Err     error
}

// ConsumeByOrder 记录核销请求，便于测试断言订单支付成功后确实触发优惠券消费。
func (m *MemoryCouponConsumer) ConsumeByOrder(_ context.Context, userID, orderID uint64) error {
	m.Calls++
	m.UserID = userID
	m.OrderID = orderID
	return m.Err
}

// LockByOrder 内存版锁定实现，仅记录调用便于断言。
func (m *MemoryCouponConsumer) LockByOrder(_ context.Context, userID, couponID, orderID uint64) error {
	m.UserID = userID
	m.OrderID = orderID
	return m.Err
}

// ReleaseByOrder 内存版释放实现，仅记录调用便于断言。
func (m *MemoryCouponConsumer) ReleaseByOrder(_ context.Context, userID, orderID uint64) error {
	m.UserID = userID
	m.OrderID = orderID
	return m.Err
}

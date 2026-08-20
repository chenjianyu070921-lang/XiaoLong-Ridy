package repository

import "context"

// MemoryCouponConsumer is a test double for coupon consumption.
type MemoryCouponConsumer struct {
	Calls   int
	UserID  uint64
	OrderID uint64
	Err     error
}

func (m *MemoryCouponConsumer) ConsumeByOrder(_ context.Context, userID, orderID uint64) error {
	m.Calls++
	m.UserID = userID
	m.OrderID = orderID
	return m.Err
}

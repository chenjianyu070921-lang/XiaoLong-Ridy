package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	userCouponStatusUsed   int8 = 2
	userCouponStatusLocked int8 = 4
)

// CouponConsumer consumes a coupon locked by the paid order.
type CouponConsumer interface {
	ConsumeByOrder(ctx context.Context, userID, orderID uint64) error
}

type gormCouponConsumer struct {
	db *gorm.DB
}

func NewGormCouponConsumer(db *gorm.DB) CouponConsumer {
	return &gormCouponConsumer{db: db}
}

// ConsumeByOrder is idempotent: it only consumes coupons still locked by orderID.
func (c *gormCouponConsumer) ConsumeByOrder(ctx context.Context, userID, orderID uint64) error {
	now := time.Now()
	return c.db.WithContext(ctx).
		Table("user_coupon").
		Where("user_id = ? AND locked_order_id = ? AND status = ?", userID, orderID, userCouponStatusLocked).
		Updates(map[string]interface{}{
			"status":          userCouponStatusUsed,
			"order_id":        orderID,
			"locked_order_id": 0,
			"locked_at":       nil,
			"used_at":         now,
			"updated_at":      now,
		}).Error
}

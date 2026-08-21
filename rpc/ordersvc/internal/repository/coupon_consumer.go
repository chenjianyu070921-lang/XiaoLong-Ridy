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

// CouponConsumer 定义订单支付成功后核销用户优惠券的最小仓储契约。
type CouponConsumer interface {
	ConsumeByOrder(ctx context.Context, userID, orderID uint64) error
}

// gormCouponConsumer 通过 user_coupon 表核销订单锁定的优惠券。
type gormCouponConsumer struct {
	db *gorm.DB
}

// NewGormCouponConsumer 创建订单服务使用的优惠券核销仓储。
func NewGormCouponConsumer(db *gorm.DB) CouponConsumer {
	return &gormCouponConsumer{db: db}
}

// ConsumeByOrder 将 locked_order_id 等于订单 ID 的用户券置为已使用。
// 若订单未使用优惠券则不报错；若支付回调重复到达且已核销同一订单，也保持幂等。
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

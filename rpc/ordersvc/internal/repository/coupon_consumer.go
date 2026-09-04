package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrCouponNotConsumed 优惠券核销失败，用于上层感知支付后核销异常。
	ErrCouponNotConsumed = errors.New("coupon consume failed")
	// ErrCouponLockFailed 优惠券锁定失败（已被占用或不存在）。
	ErrCouponLockFailed = errors.New("coupon lock failed")
)

const (
	userCouponStatusAvailable int8 = 1
	userCouponStatusUsed      int8 = 2
	userCouponStatusLocked    int8 = 4
)

// CouponConsumer 定义订单支付成功后核销用户优惠券的最小仓储契约。
type CouponConsumer interface {
	// LockByOrder 下单时锁定一张可用优惠券到指定订单，避免并发重复占用。
	LockByOrder(ctx context.Context, userID, couponID, orderID uint64) error
	ConsumeByOrder(ctx context.Context, userID, orderID uint64) error
	// ReleaseByOrder 释放订单锁定的优惠券（取消/退款时回滚），幂等。
	ReleaseByOrder(ctx context.Context, userID, orderID uint64) error
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

// ReleaseByOrder 将锁定在该订单上的优惠券释放回可用状态（取消/退款时回滚）。
func (c *gormCouponConsumer) ReleaseByOrder(ctx context.Context, userID, orderID uint64) error {
	return c.db.WithContext(ctx).
		Table("user_coupon").
		Where("user_id = ? AND locked_order_id = ? AND status = ?", userID, orderID, userCouponStatusLocked).
		Updates(map[string]interface{}{
			"status":          userCouponStatusAvailable,
			"locked_order_id": 0,
			"locked_at":       nil,
			"updated_at":      time.Now(),
		}).Error
}

// LockByOrder 下单时把指定可用券绑定到订单并置为锁定，CAS 防止并发重复占用。
func (c *gormCouponConsumer) LockByOrder(ctx context.Context, userID, couponID, orderID uint64) error {
	res := c.db.WithContext(ctx).
		Table("user_coupon").
		Where("id = ? AND user_id = ? AND status = ?", couponID, userID, userCouponStatusAvailable).
		Updates(map[string]interface{}{
			"status":          userCouponStatusLocked,
			"locked_order_id": orderID,
			"locked_at":       time.Now(),
			"updated_at":      time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCouponLockFailed
	}
	return nil
}

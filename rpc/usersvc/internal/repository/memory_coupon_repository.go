package repository

import (
	"context"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

// MemoryCouponRepository 是本地开发和测试使用的优惠券内存仓储。
type MemoryCouponRepository struct {
	mu          sync.RWMutex
	nextID      uint64
	coupons     map[uint64]*model.Coupon
	userCoupons map[uint64]*model.UserCoupon
}

// NewMemoryCouponRepository 创建优惠券内存仓储。
func NewMemoryCouponRepository() *MemoryCouponRepository {
	return &MemoryCouponRepository{
		nextID:      1,
		coupons:     make(map[uint64]*model.Coupon),
		userCoupons: make(map[uint64]*model.UserCoupon),
	}
}

// AddCouponForTest 写入测试用优惠券模板。
func (r *MemoryCouponRepository) AddCouponForTest(coupon *model.Coupon) {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *coupon
	r.coupons[copied.ID] = &copied
}

// Claim 领取优惠券，并维护库存和单用户领取上限。
func (r *MemoryCouponRepository) Claim(_ context.Context, userID, couponID uint64) (*UserCouponWithTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	coupon, ok := r.coupons[couponID]
	if !ok {
		return nil, ErrCouponNotFound
	}
	now := time.Now()
	if !isCouponAvailable(coupon, now) {
		return nil, ErrCouponUnavailable
	}
	if coupon.TotalCount > 0 && coupon.ReceivedCount >= coupon.TotalCount {
		return nil, ErrCouponUnavailable
	}
	limit := coupon.PerUserLimit
	if limit <= 0 {
		limit = 1
	}
	received := 0
	for _, item := range r.userCoupons {
		if item.UserID == userID && item.CouponID == couponID {
			received++
		}
	}
	if received >= limit {
		return nil, ErrCouponReceiveLimit
	}

	id := r.nextID
	r.nextID++
	userCoupon := &model.UserCoupon{
		ID:         id,
		UserID:     userID,
		CouponID:   couponID,
		Status:     model.UserCouponStatusUnused,
		ReceivedAt: now,
		ExpireAt:   coupon.ValidEndAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.userCoupons[id] = userCoupon
	coupon.ReceivedCount++
	return couponView(userCoupon, coupon), nil
}

// ListByUser 查询用户自己的优惠券列表，status 为 0 时返回全部状态。
func (r *MemoryCouponRepository) ListByUser(_ context.Context, userID uint64, status int8) ([]*UserCouponWithTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	list := make([]*UserCouponWithTemplate, 0)
	for _, item := range r.userCoupons {
		if item.UserID != userID {
			continue
		}
		current := *item
		if current.Status == model.UserCouponStatusUnused && current.ExpireAt.Before(now) {
			current.Status = model.UserCouponStatusExpired
		}
		if status > 0 && current.Status != status {
			continue
		}
		coupon, ok := r.coupons[current.CouponID]
		if !ok {
			continue
		}
		list = append(list, couponView(&current, coupon))
	}
	return list, nil
}

// Lock 将用户券从未使用锁定到指定订单，防止重复下单使用。
func (r *MemoryCouponRepository) Lock(_ context.Context, userID, userCouponID, orderID uint64, carType int8, cityCode string) (*UserCouponWithTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userCoupon, ok := r.userCoupons[userCouponID]
	if !ok || userCoupon.UserID != userID {
		return nil, ErrUserCouponNotFound
	}
	coupon, ok := r.coupons[userCoupon.CouponID]
	if !ok {
		return nil, ErrCouponNotFound
	}
	now := time.Now()
	if !isCouponUsable(userCoupon, coupon, now, carType, cityCode) {
		return nil, ErrCouponUnavailable
	}
	userCoupon.Status = model.UserCouponStatusLocked
	userCoupon.LockedOrderID = orderID
	userCoupon.LockedAt = &now
	userCoupon.UpdatedAt = now
	return couponView(userCoupon, coupon), nil
}

// Release 将指定订单锁定的用户券释放回未使用状态。
func (r *MemoryCouponRepository) Release(_ context.Context, userID, userCouponID, orderID uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userCoupon, ok := r.userCoupons[userCouponID]
	if !ok || userCoupon.UserID != userID {
		return ErrUserCouponNotFound
	}
	if userCoupon.Status != model.UserCouponStatusLocked || userCoupon.LockedOrderID != orderID {
		return ErrCouponUnavailable
	}
	now := time.Now()
	userCoupon.Status = model.UserCouponStatusUnused
	userCoupon.LockedOrderID = 0
	userCoupon.LockedAt = nil
	userCoupon.UpdatedAt = now
	return nil
}

// isCouponAvailable 判断优惠券模板是否处于可领取窗口。
// ConsumeByOrder 将当前订单锁定的用户券核销为已使用状态，支付成功后由订单服务调用。
func (r *MemoryCouponRepository) ConsumeByOrder(_ context.Context, userID, orderID uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, userCoupon := range r.userCoupons {
		if userCoupon.UserID != userID || userCoupon.LockedOrderID != orderID {
			continue
		}
		if userCoupon.Status == model.UserCouponStatusUsed && userCoupon.OrderID == orderID {
			return nil
		}
		if userCoupon.Status != model.UserCouponStatusLocked {
			return ErrCouponUnavailable
		}
		now := time.Now()
		userCoupon.Status = model.UserCouponStatusUsed
		userCoupon.OrderID = orderID
		userCoupon.LockedOrderID = 0
		userCoupon.LockedAt = nil
		userCoupon.UsedAt = &now
		userCoupon.UpdatedAt = now
		return nil
	}
	return nil
}

func isCouponAvailable(coupon *model.Coupon, now time.Time) bool {
	return coupon != nil &&
		coupon.Status == model.CouponStatusEnabled &&
		!now.Before(coupon.ValidStartAt) &&
		now.Before(coupon.ValidEndAt)
}

// couponView 复制用户券和模板，避免调用方修改仓储内部状态。
func couponView(userCoupon *model.UserCoupon, coupon *model.Coupon) *UserCouponWithTemplate {
	userCouponCopy := *userCoupon
	couponCopy := *coupon
	return &UserCouponWithTemplate{
		UserCoupon: &userCouponCopy,
		Coupon:     &couponCopy,
	}
}

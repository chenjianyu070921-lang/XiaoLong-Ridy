package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

var (
	// ErrCouponNotFound 表示优惠券模板不存在。
	ErrCouponNotFound = errors.New("coupon not found")
	// ErrCouponUnavailable 表示优惠券未启用、过期或库存不足。
	ErrCouponUnavailable = errors.New("coupon unavailable")
	// ErrCouponReceiveLimit 表示用户领取次数已达到模板限制。
	ErrCouponReceiveLimit = errors.New("coupon receive limit exceeded")
	// ErrUserCouponNotFound 表示用户优惠券不存在或不属于当前用户。
	ErrUserCouponNotFound = errors.New("user coupon not found")
)

// UserCouponWithTemplate 聚合用户券实例和优惠券模板，便于列表展示。
type UserCouponWithTemplate struct {
	UserCoupon *model.UserCoupon
	Coupon     *model.Coupon
}

// CouponRepository 定义优惠券领取和查询仓储契约。
type CouponRepository interface {
	Claim(ctx context.Context, userID, couponID uint64) (*UserCouponWithTemplate, error)
	ListByUser(ctx context.Context, userID uint64, status int8) ([]*UserCouponWithTemplate, error)
	ListByUserPage(ctx context.Context, userID uint64, status int8, page, pageSize int) ([]*UserCouponWithTemplate, int64, error)
	Lock(ctx context.Context, userID, userCouponID, orderID uint64, carType int8, cityCode string) (*UserCouponWithTemplate, error)
	Release(ctx context.Context, userID, userCouponID, orderID uint64) error
	ConsumeByOrder(ctx context.Context, userID, orderID uint64) error
}

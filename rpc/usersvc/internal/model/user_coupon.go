package model

import "time"

const (
	// UserCouponStatusUnused 表示用户优惠券尚未使用。
	UserCouponStatusUnused int8 = 1
	// UserCouponStatusUsed 表示用户优惠券已经完成核销。
	UserCouponStatusUsed int8 = 2
	// UserCouponStatusExpired 表示用户优惠券已经超过有效期。
	UserCouponStatusExpired int8 = 3
	// UserCouponStatusLocked 表示用户优惠券已被下单流程临时锁定。
	UserCouponStatusLocked int8 = 4
)

// UserCoupon 对应 user_coupon 表，保存乘客领取的优惠券实例及其锁定、核销状态。
type UserCoupon struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID        uint64     `gorm:"column:user_id;not null;index:idx_user_status,priority:1" json:"userId"`
	CouponID      uint64     `gorm:"column:coupon_id;not null;index:idx_coupon_id" json:"couponId"`
	OrderID       uint64     `gorm:"column:order_id;not null;default:0" json:"orderId"`
	LockedOrderID uint64     `gorm:"column:locked_order_id;not null;default:0;index:idx_locked_order" json:"lockedOrderId"`
	Status        int8       `gorm:"column:status;not null;default:1;index:idx_user_status,priority:2;index:idx_expire_status,priority:1" json:"status"`
	ReceivedAt    time.Time  `gorm:"column:received_at;not null;autoCreateTime" json:"receivedAt"`
	UsedAt        *time.Time `gorm:"column:used_at" json:"usedAt"`
	LockedAt      *time.Time `gorm:"column:locked_at" json:"lockedAt"`
	ExpireAt      time.Time  `gorm:"column:expire_at;not null;index:idx_expire_status,priority:2" json:"expireAt"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回用户优惠券模型对应的数据表名称，供 GORM 映射 user_coupon 表。
func (UserCoupon) TableName() string {
	return "user_coupon"
}

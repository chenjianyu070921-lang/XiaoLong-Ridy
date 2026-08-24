package model

import "time"

const (
	// CouponTypeFullReduction 表示满足金额门槛后减免固定金额。
	CouponTypeFullReduction int8 = 1
	// CouponTypeDiscount 表示按折扣率减免订单金额。
	CouponTypeDiscount int8 = 2
	// CouponTypeInstantReduction 表示无门槛减免固定金额。
	CouponTypeInstantReduction int8 = 3

	// CouponCarTypeAll 表示优惠券适用于全部车型。
	CouponCarTypeAll int8 = 0
	// CouponCarTypeEconomy 表示优惠券适用于特惠快车。
	CouponCarTypeEconomy int8 = 1
	// CouponCarTypeExpress 表示优惠券适用于快车。
	CouponCarTypeExpress int8 = 2
	// CouponCarTypeCarpool 表示优惠券适用于拼车。
	CouponCarTypeCarpool int8 = 3

	// CouponStatusEnabled 表示优惠券模板已启用。
	CouponStatusEnabled int8 = 2
	// CouponStatusDisabled 表示优惠券模板已停用。
	CouponStatusDisabled int8 = 3
)

// Coupon 对应 coupon 表，保存优惠券模板、使用门槛和适用范围。
type Coupon struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name            string    `gorm:"column:name;size:50;not null" json:"name"`
	Type            int8      `gorm:"column:type;not null;default:1" json:"type"`
	FaceValue       float64   `gorm:"column:face_value;type:decimal(10,2);not null;default:0.00" json:"faceValue"`
	Discount        float64   `gorm:"column:discount;type:decimal(3,2);not null;default:1.00" json:"discount"`
	ThresholdAmount float64   `gorm:"column:threshold_amount;type:decimal(10,2);not null;default:0.00" json:"thresholdAmount"`
	CarType         int8      `gorm:"column:car_type;not null;default:0;index:idx_car_city,priority:1" json:"carType"`
	CityCode        string    `gorm:"column:city_code;size:20;not null;default:'';index:idx_car_city,priority:2" json:"cityCode"`
	TotalCount      int       `gorm:"column:total_count;not null;default:0" json:"totalCount"`
	ReceivedCount   int       `gorm:"column:received_count;not null;default:0" json:"receivedCount"`
	PerUserLimit    int       `gorm:"column:per_user_limit;not null;default:1" json:"perUserLimit"`
	ValidStartAt    time.Time `gorm:"column:valid_start_at;not null" json:"validStartAt"`
	ValidEndAt      time.Time `gorm:"column:valid_end_at;not null" json:"validEndAt"`
	Status          int8      `gorm:"column:status;not null;default:1;index:idx_status;index:idx_car_city,priority:3" json:"status"`
	Remark          string    `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

// TableName 返回优惠券模板模型对应的数据表名称，供 GORM 映射 coupon 表。
func (Coupon) TableName() string {
	return "coupon"
}

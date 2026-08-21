package model

import "time"

// RideOrder 对应 ride_order 表：打车订单主表。
type RideOrder struct {
	Id                 uint64     `gorm:"primaryKey;column:id" json:"id"`
	OrderNo            string     `gorm:"column:order_no;size:32" json:"orderNo"`
	UserId             uint64     `gorm:"column:user_id" json:"userId"`
	DriverId           uint64     `gorm:"column:driver_id;default:0" json:"driverId"`
	CarType            int8       `gorm:"column:car_type;default:1" json:"carType"`
	CityCode           string     `gorm:"column:city_code;size:16;default:''" json:"cityCode"`
	FromAddress        string     `gorm:"column:from_address;size:255" json:"fromAddress"`
	FromLongitude      float64    `gorm:"column:from_longitude;type:decimal(10,6)" json:"fromLongitude"`
	FromLatitude       float64    `gorm:"column:from_latitude;type:decimal(10,6)" json:"fromLatitude"`
	ToAddress          string     `gorm:"column:to_address;size:255" json:"toAddress"`
	ToLongitude        float64    `gorm:"column:to_longitude;type:decimal(10,6)" json:"toLongitude"`
	ToLatitude         float64    `gorm:"column:to_latitude;type:decimal(10,6)" json:"toLatitude"`
	EstimatedDistanceM int        `gorm:"column:estimated_distance_m;default:0" json:"estimatedDistanceM"`
	EstimatedDurationS int        `gorm:"column:estimated_duration_s;default:0" json:"estimatedDurationS"`
	EstimatedPrice     float64    `gorm:"column:estimated_price;type:decimal(10,2);default:0" json:"estimatedPrice"`
	Status             int8       `gorm:"column:status;default:1" json:"status"`
	CancelReason       string     `gorm:"column:cancel_reason;size:255;default:''" json:"cancelReason"`
	CancelBy           string     `gorm:"column:cancel_by;size:20;default:''" json:"cancelBy"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt          *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

func (RideOrder) TableName() string {
	return "ride_order"
}

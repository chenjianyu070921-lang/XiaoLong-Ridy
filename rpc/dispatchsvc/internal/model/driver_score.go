package model

import "time"

// DriverScore 对应 driver_score 表：司机服务质量分。
type DriverScore struct {
	Id                  uint64    `gorm:"primaryKey;column:id" json:"id"`
	DriverId            uint64    `gorm:"column:driver_id" json:"driverId"`
	Score               float64   `gorm:"column:score" json:"score"`
	Level               int8      `gorm:"column:level" json:"level"`
	MonthOrders         int       `gorm:"column:month_orders" json:"monthOrders"`
	MonthCancelRate     float64   `gorm:"column:month_cancel_rate" json:"monthCancelRate"`
	MonthComplaintCount int       `gorm:"column:month_complaint_count" json:"monthComplaintCount"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (DriverScore) TableName() string {
	return "driver_score"
}

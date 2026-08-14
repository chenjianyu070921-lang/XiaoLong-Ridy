package model

import "time"

// DriverScore 对应 driver_score 表：司机服务分、等级与运营指标。
type DriverScore struct {
	Id                 uint64    `gorm:"primaryKey;column:id" json:"id"`
	DriverId           uint64    `gorm:"column:driver_id" json:"driverId"`
	Score              float64   `gorm:"column:score;default:100.00" json:"score"`
	Level              int8      `gorm:"column:level;default:1" json:"level"`
	MonthOrders        int       `gorm:"column:month_orders;default:0" json:"monthOrders"`
	MonthCancelRate    float64   `gorm:"column:month_cancel_rate;default:0.00" json:"monthCancelRate"`
	MonthComplaintCount int      `gorm:"column:month_complaint_count;default:0" json:"monthComplaintCount"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 返回对应的数据库表名。
func (DriverScore) TableName() string {
	return "driver_score"
}

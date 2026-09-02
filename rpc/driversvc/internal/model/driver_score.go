package model

import "time"

// DriverScore 对应 driver_score 表：司机服务分、等级与运营指标。
type DriverScore struct {
	Id                  uint64    `gorm:"primaryKey;column:id" json:"id"`                                    // 评分记录 ID
	DriverId            uint64    `gorm:"column:driver_id;not null;uniqueIndex:uk_driver_id" json:"driverId"` // 司机 ID
	Score               float64   `gorm:"column:score;default:100.00" json:"score"`                          // 服务分
	Level               int8      `gorm:"column:level;default:1" json:"level"`                               // 司机等级
	MonthOrders         int       `gorm:"column:month_orders;default:0" json:"monthOrders"`                  // 本月完成订单数
	MonthCancelRate     float64   `gorm:"column:month_cancel_rate;default:0.00" json:"monthCancelRate"`      // 本月取消率
	MonthComplaintCount int       `gorm:"column:month_complaint_count;default:0" json:"monthComplaintCount"` // 本月投诉数
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updatedAt"`                                // 更新时间
}

// TableName 返回对应的数据库表名。
func (DriverScore) TableName() string {
	return "driver_score"
}

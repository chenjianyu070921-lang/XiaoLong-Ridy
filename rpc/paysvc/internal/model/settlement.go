package model

import "time"

// 结算状态
const (
	SettlementStatusPending = 1 // 待结算
	SettlementStatusSettled = 2 // 已结算
)

// Settlement 对应 settlement 表：订单结算。
type Settlement struct {
	Id                     uint64     `gorm:"primaryKey;column:id" json:"id"`
	SettlementNo           string     `gorm:"column:settlement_no;size:32" json:"settlementNo"`
	OrderId                uint64     `gorm:"column:order_id" json:"orderId"`
	DriverId               uint64     `gorm:"column:driver_id" json:"driverId"`
	TotalAmount            float64    `gorm:"column:total_amount;type:decimal(10,2)" json:"totalAmount"`
	PlatformCommissionRate float64    `gorm:"column:platform_commission_rate;type:decimal(5,2);default:0" json:"platformCommissionRate"`
	PlatformCommission     float64    `gorm:"column:platform_commission;type:decimal(10,2);default:0" json:"platformCommission"`
	DriverIncome           float64    `gorm:"column:driver_income;type:decimal(10,2);default:0" json:"driverIncome"`
	Status                 int8       `gorm:"column:status;default:1" json:"status"`
	AutoSettled            bool       `gorm:"column:auto_settled;default:0" json:"autoSettled"`
	SettledJobRunID        string     `gorm:"column:settled_job_run_id;size:32" json:"settledJobRunId"`
	SettledAt              *time.Time `gorm:"column:settled_at" json:"settledAt"`
	CreatedAt              time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (Settlement) TableName() string {
	return "settlement"
}

package model

import "time"

// DriverWithdraw 对应 driver_withdraw 表：司机提现申请与打款结果。
type DriverWithdraw struct {
	Id         uint64     `gorm:"primaryKey;column:id" json:"id"`
	DriverId   uint64     `gorm:"column:driver_id" json:"driverId"`
	WithdrawNo string     `gorm:"column:withdraw_no;size:32" json:"withdrawNo"`
	Amount     float64    `gorm:"column:amount" json:"amount"`
	PayeeName  string     `gorm:"column:payee_name;size:50;default:''" json:"payeeName"`
	PayAccount string     `gorm:"column:pay_account;size:100;default:''" json:"payAccount"`
	Status     int8       `gorm:"column:status;default:1" json:"status"`
	Remark     string     `gorm:"column:remark;size:255;default:''" json:"remark"`
	AppliedAt  *time.Time `gorm:"column:applied_at" json:"appliedAt"`
	PaidAt     *time.Time `gorm:"column:paid_at" json:"paidAt"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"createdAt"`
}

// TableName 返回对应的数据库表名。
func (DriverWithdraw) TableName() string {
	return "driver_withdraw"
}

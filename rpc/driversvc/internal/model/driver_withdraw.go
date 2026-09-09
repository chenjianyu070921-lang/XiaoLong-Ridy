package model

import "time"

// DriverWithdraw 对应 driver_withdraw 表：司机提现申请与打款结果。
type DriverWithdraw struct {
	Id         uint64     `gorm:"primaryKey;column:id" json:"id"`                           // 提现记录 ID
	DriverId   uint64     `gorm:"column:driver_id" json:"driverId"`                         // 司机 ID
	WithdrawNo string     `gorm:"column:withdraw_no;size:32" json:"withdrawNo"`             // 提现单号
	Amount     float64    `gorm:"column:amount" json:"amount"`                              // 提现金额
	PayeeName  string     `gorm:"column:payee_name;size:50;default:''" json:"payeeName"`    // 收款人姓名
	PayAccount string     `gorm:"column:pay_account;size:100;default:''" json:"payAccount"` // 收款账户
	Status     int8       `gorm:"column:status;default:1" json:"status"`                    // 状态：1申请中、2处理中、3已完成、4失败
	Remark     string     `gorm:"column:remark;size:255;default:''" json:"remark"`          // 处理备注
	AppliedAt  *time.Time `gorm:"column:applied_at" json:"appliedAt"`                       // 申请时间
	PaidAt     *time.Time `gorm:"column:paid_at" json:"paidAt"`                             // 打款时间
	CreatedAt  time.Time  `gorm:"column:created_at" json:"createdAt"`                       // 创建时间
}

// TableName 返回对应的数据库表名。
func (DriverWithdraw) TableName() string {
	return "driver_withdraw"
}

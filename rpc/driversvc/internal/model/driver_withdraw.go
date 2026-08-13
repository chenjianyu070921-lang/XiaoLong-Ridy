package model

import "time"

// DriverWithdraw 对应 driver_withdraw 表：司机提现申请与打款结果。
type DriverWithdraw struct {
	Id          uint64     `gorm:"primaryKey;column:id" json:"id"`                      // Id：提现主键 ID（自增）
	DriverId    uint64     `gorm:"column:driver_id" json:"driverId"`                    // DriverId：所属司机 ID
	WithdrawNo  string     `gorm:"column:withdraw_no;size:32" json:"withdrawNo"`        // WithdrawNo：提现单号，唯一
	Amount      float64    `gorm:"column:amount" json:"amount"`                         // Amount：提现金额
	PayeeName   string     `gorm:"column:payee_name;size:50;default:''" json:"payeeName"`     // PayeeName：收款人姓名
	PayAccount  string     `gorm:"column:pay_account;size:100;default:''" json:"payAccount"` // PayAccount：收款账号
	Status      int8       `gorm:"column:status;default:1" json:"status"`              // Status：状态；1申请中 2打款成功 3打款失败
	Remark      string     `gorm:"column:remark;size:255;default:''" json:"remark"`    // Remark：失败原因/备注
	AppliedAt   *time.Time `gorm:"column:applied_at" json:"appliedAt"`                 // AppliedAt：申请时间
	PaidAt      *time.Time `gorm:"column:paid_at" json:"paidAt"`                       // PaidAt：打款时间
	CreatedAt   time.Time  `gorm:"column:created_at" json:"createdAt"`                 // CreatedAt：创建时间
}

// TableName 返回对应的数据库表名。
func (DriverWithdraw) TableName() string {
	return "driver_withdraw"
}

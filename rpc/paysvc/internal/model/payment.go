package model

import "time"

// 支付状态
const (
	PaymentStatusPending = 1 // 待支付
	PaymentStatusPaid    = 2 // 支付成功
	PaymentStatusFailed  = 3 // 支付失败
	PaymentStatusRefund  = 4 // 已退款
)

// Payment 对应 payment 表：支付单。
type Payment struct {
	Id            uint64     `gorm:"primaryKey;column:id" json:"id"`
	PaymentNo     string     `gorm:"column:payment_no;size:32" json:"paymentNo"`
	OrderId       uint64     `gorm:"column:order_id" json:"orderId"`
	UserId        uint64     `gorm:"column:user_id" json:"userId"`
	Amount        float64    `gorm:"column:amount;type:decimal(10,2)" json:"amount"`
	Channel       string     `gorm:"column:channel;size:20;default:wechat" json:"channel"`
	Status        int8       `gorm:"column:status;default:1" json:"status"`
	TransactionId string     `gorm:"column:transaction_id;size:64;default:''" json:"transactionId"`
	RefundAmount  float64    `gorm:"column:refund_amount;type:decimal(10,2);default:0" json:"refundAmount"`
	PaidAt        *time.Time `gorm:"column:paid_at" json:"paidAt"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (Payment) TableName() string {
	return "payment"
}

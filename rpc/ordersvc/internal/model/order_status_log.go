package model

import "time"

// OrderStatusLog 对应 order_status_log 表：订单状态流转日志。
type OrderStatusLog struct {
	Id           uint64 `gorm:"primaryKey;column:id" json:"id"`
	OrderId      uint64 `gorm:"column:order_id" json:"orderId"`
	FromStatus   int8   `gorm:"column:from_status" json:"fromStatus"`
	ToStatus     int8   `gorm:"column:to_status" json:"toStatus"`
	OperatorType string `gorm:"column:operator_type;size:20;default:system" json:"operatorType"`
	OperatorId   uint64 `gorm:"column:operator_id;default:0" json:"operatorId"`
	Remark       string `gorm:"column:remark;size:255;default:''" json:"remark"`
	// PaidCents 仅用于支付完成事务向订单仓储传递实付金额，不落库到日志表。
	PaidCents int64     `gorm:"-" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (OrderStatusLog) TableName() string {
	return "order_status_log"
}

package model

import "time"

// DispatchRecord 对应 dispatch_record 表：订单向候选司机的派单记录。
type DispatchRecord struct {
	Id           uint64    `gorm:"primaryKey;column:id" json:"id"`
	OrderId      uint64    `gorm:"column:order_id" json:"orderId"`
	DriverId     uint64    `gorm:"column:driver_id" json:"driverId"`
	DispatchType int8      `gorm:"column:dispatch_type;default:1" json:"dispatchType"`
	Status       int8      `gorm:"column:status;default:1" json:"status"`
	MatchScore   float64   `gorm:"column:match_score;type:decimal(10,2);default:0" json:"matchScore"`
	Remark       string    `gorm:"column:remark;size:255;default:''" json:"remark"`
	RejectReason string    `gorm:"column:reject_reason;size:255;default:''" json:"rejectReason"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 返回对应数据库表名。
func (DispatchRecord) TableName() string {
	return "dispatch_record"
}

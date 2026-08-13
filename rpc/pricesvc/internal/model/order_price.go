package model

import "time"

// OrderPrice 对应 order_price 表：订单价格明细（预估/实付快照）。
type OrderPrice struct {
	Id              uint64    `gorm:"primaryKey;column:id" json:"id"`
	OrderId         uint64    `gorm:"column:order_id" json:"orderId"`
	PriceRuleId     uint64    `gorm:"column:price_rule_id;default:0" json:"priceRuleId"`
	EstimatedPrice  float64   `gorm:"column:estimated_price;type:decimal(10,2)" json:"estimatedPrice"`
	ActualPrice     float64   `gorm:"column:actual_price;type:decimal(10,2);default:0" json:"actualPrice"`
	BaseFee         float64   `gorm:"column:base_fee;type:decimal(10,2);default:0" json:"baseFee"`
	DistanceFee     float64   `gorm:"column:distance_fee;type:decimal(10,2);default:0" json:"distanceFee"`
	TimeFee         float64   `gorm:"column:time_fee;type:decimal(10,2);default:0" json:"timeFee"`
	NightFee        float64   `gorm:"column:night_fee;type:decimal(10,2);default:0" json:"nightFee"`
	DynamicFee      float64   `gorm:"column:dynamic_fee;type:decimal(10,2);default:0" json:"dynamicFee"`
	DiscountAmount  float64   `gorm:"column:discount_amount;type:decimal(10,2);default:0" json:"discountAmount"`
	PlatformSubsidy float64   `gorm:"column:platform_subsidy;type:decimal(10,2);default:0" json:"platformSubsidy"`
	PayableAmount   float64   `gorm:"column:payable_amount;type:decimal(10,2);default:0" json:"payableAmount"`
	Status          int8      `gorm:"column:status;default:1" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (OrderPrice) TableName() string {
	return "order_price"
}

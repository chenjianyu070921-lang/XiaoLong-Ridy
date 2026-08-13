package model

import "time"

// PriceRule 对应 price_rule 表：计价规则。
type PriceRule struct {
	Id              uint64     `gorm:"primaryKey;column:id" json:"id"`
	Name            string     `gorm:"column:name;size:50" json:"name"`
	CityCode        string     `gorm:"column:city_code;size:20;default:''" json:"cityCode"`
	CarType         int8       `gorm:"column:car_type;default:1" json:"carType"`
	BasePrice       float64    `gorm:"column:base_price;type:decimal(10,2)" json:"basePrice"`
	BaseDistanceKm  float64    `gorm:"column:base_distance_km;type:decimal(6,2);default:0" json:"baseDistanceKm"`
	PerKmPrice      float64    `gorm:"column:per_km_price;type:decimal(6,2)" json:"perKmPrice"`
	PerMinutePrice  float64    `gorm:"column:per_minute_price;type:decimal(6,2)" json:"perMinutePrice"`
	NightStartTime  *string    `gorm:"column:night_start_time;type:time" json:"nightStartTime"`
	NightEndTime    *string    `gorm:"column:night_end_time;type:time" json:"nightEndTime"`
	NightSurcharge  float64    `gorm:"column:night_surcharge;type:decimal(10,2);default:0" json:"nightSurcharge"`
	DynamicMaxFactor float64   `gorm:"column:dynamic_max_factor;type:decimal(3,2);default:1" json:"dynamicMaxFactor"`
	Status          int8       `gorm:"column:status;default:1" json:"status"`
	EffectiveAt     time.Time  `gorm:"column:effective_at" json:"effectiveAt"`
	ExpireAt        *time.Time `gorm:"column:expire_at" json:"expireAt"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (PriceRule) TableName() string {
	return "price_rule"
}

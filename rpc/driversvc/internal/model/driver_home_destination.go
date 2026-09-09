package model

import "time"

// DriverHomeDestination 司机回家目的地与顺路回家模式设置。
type DriverHomeDestination struct {
	ID             uint64    `gorm:"primaryKey;column:id"`
	DriverID       uint64    `gorm:"column:driver_id"`
	HomeAddr       string    `gorm:"column:home_addr"`
	HomeLng        float64   `gorm:"column:home_lng"`
	HomeLat        float64   `gorm:"column:home_lat"`
	IsHomeModeOpen int8      `gorm:"column:is_home_mode_open"`
	MaxDetourRatio float64   `gorm:"column:max_detour_ratio"`
	CreateAt       time.Time `gorm:"column:create_at"`
	UpdateAt       time.Time `gorm:"column:update_at"`
}

// TableName 返回回家目的地设置表名。
func (DriverHomeDestination) TableName() string {
	return "driver_home_destination"
}

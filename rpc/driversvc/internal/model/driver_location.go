package model

import "time"

// DriverLocation 对应 driver_location 表：保存司机当前最新位置和在线状态。
type DriverLocation struct {
	Id           uint64    `gorm:"primaryKey;column:id" json:"id"`           // 位置记录 ID
	DriverID     uint64    `gorm:"column:driver_id" json:"driverId"`         // 司机 ID
	Longitude    float64   `gorm:"column:longitude" json:"longitude"`        // 经度
	Latitude     float64   `gorm:"column:latitude" json:"latitude"`          // 纬度
	Heading      int16     `gorm:"column:heading" json:"heading"`            // 行驶方向角（度）
	SpeedKmh     float64   `gorm:"column:speed_kmh" json:"speedKmh"`         // 速度（公里/小时）
	OnlineStatus int8      `gorm:"column:online_status" json:"onlineStatus"` // 在线状态：0离线、1在线、2行程中
	ReportTime   time.Time `gorm:"column:report_time" json:"reportTime"`     // 位置上报时间
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`       // 创建时间
	// DistanceMeters 是查询时按 Haversine 计算出的距中心点距离（非表列，仅查询结果映射用，标记为只读避免写入时报未知列）。
	DistanceMeters float64 `gorm:"column:distance_meters;->" json:"distanceMeters"`
}

// TableName 返回对应的数据库表名。
func (DriverLocation) TableName() string {
	return "driver_location"
}

// 在线状态常量，与 driver_location.online_status 字段语义一致。
const (
	// LocationOffline 表示司机离线（不接单）。
	LocationOffline int8 = 0
	// LocationOnline 表示司机在线（可接单）。
	LocationOnline int8 = 1
	// LocationOnTrip 表示司机行程中。
	LocationOnTrip int8 = 2
)

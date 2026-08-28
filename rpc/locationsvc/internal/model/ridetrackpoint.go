package model

import (
	"time"

	"gorm.io/gorm"
)

// RideTrackPoint 表示订单行程轨迹点，对应 scripts/sql/migrate/06_location_module.sql 中的 ride_track_point 表。
// 本模型只使用既有表结构，不参与 AutoMigrate，避免位置服务启动时擅自修改数据库结构。
type RideTrackPoint struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID    uint64    `gorm:"column:order_id;not null"`
	DriverID   uint64    `gorm:"column:driver_id;not null"`
	Longitude  float64   `gorm:"column:longitude;type:decimal(10,6);not null"`
	Latitude   float64   `gorm:"column:latitude;type:decimal(10,6);not null"`
	SpeedKmh   float64   `gorm:"column:speed_kmh;type:decimal(5,1);default:0"`
	Direction  int16     `gorm:"column:direction;type:smallint;default:0"`
	RecordedAt time.Time `gorm:"column:recorded_at;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回订单轨迹点表名。
func (RideTrackPoint) TableName() string {
	return "ride_track_point"
}

// RideTrackPointModel 封装订单轨迹点读写。
type RideTrackPointModel struct {
	db *gorm.DB
}

// NewRideTrackPointModel 创建订单轨迹点数据访问对象。
func NewRideTrackPointModel(db *gorm.DB) *RideTrackPointModel {
	return &RideTrackPointModel{db: db}
}

// Insert 写入单个订单轨迹点，供司机行程中位置上报时落库。
func (m *RideTrackPointModel) Insert(point *RideTrackPoint) error {
	return m.db.Create(point).Error
}

// ListByOrder 按订单 ID 和可选时间范围查询轨迹点，按上报时间正序返回，供后台和客服回放使用。
func (m *RideTrackPointModel) ListByOrder(orderID int64, startTime, endTime time.Time, limit int) ([]RideTrackPoint, error) {
	query := m.db.Where("order_id = ?", orderID)
	if !startTime.IsZero() {
		query = query.Where("recorded_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("recorded_at <= ?", endTime)
	}
	var points []RideTrackPoint
	err := query.Order("recorded_at ASC, id ASC").Limit(limit).Find(&points).Error
	return points, err
}

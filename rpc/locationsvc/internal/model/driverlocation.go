package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DriverLocation 司机实时位置表（对应 scripts/sql/migrate/06_location_module.sql 的 driver_location 表）
type DriverLocation struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	DriverID     uint64    `gorm:"column:driver_id;not null;uniqueIndex:uk_driver_id"`
	Longitude    float64   `gorm:"column:longitude;type:decimal(10,6);not null"`
	Latitude     float64   `gorm:"column:latitude;type:decimal(10,6);not null"`
	Heading      int16     `gorm:"column:heading;type:smallint;default:0"`
	SpeedKmh     float64   `gorm:"column:speed_kmh;type:decimal(5,1);default:0"`
	OnlineStatus int8      `gorm:"column:online_status;type:tinyint;default:0"`
	ReportTime   time.Time `gorm:"column:report_time;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (DriverLocation) TableName() string {
	return "driver_location"
}

// DriverLocationModel 司机位置数据访问层
type DriverLocationModel struct {
	db *gorm.DB
}

func NewDriverLocationModel(db *gorm.DB) *DriverLocationModel {
	return &DriverLocationModel{db: db}
}

// Upsert 写入司机最新位置：driver_id 已存在则更新最新位置字段
func (m *DriverLocationModel) Upsert(loc *DriverLocation) error {
	return m.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "driver_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"longitude", "latitude", "heading", "speed_kmh", "online_status", "report_time",
		}),
	}).Create(loc).Error
}

// GetByDriverID 查询司机最新位置，未找到时返回 gorm.ErrRecordNotFound。
func (m *DriverLocationModel) GetByDriverID(driverID uint64) (*DriverLocation, error) {
	var loc DriverLocation
	if err := m.db.Where("driver_id = ?", driverID).First(&loc).Error; err != nil {
		return nil, err
	}
	return &loc, nil
}

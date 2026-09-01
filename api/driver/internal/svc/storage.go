package svc

import (
	"context"
	"math"
	"time"

	"XiaoLong-Ridy/common/constants"

	"gorm.io/gorm"
)

const maxTrajectoryPoints = 2000

type PassengerReviewRecord struct {
	OrderID   int64
	Rating    int32
	Comment   string
	CreatedAt time.Time
}

type ReviewRepository interface {
	ListByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]PassengerReviewRecord, int64, error)
}

type TrajectoryRecord struct {
	OrderID    int64
	DriverID   int64
	Longitude  float64
	Latitude   float64
	SpeedKmh   float64
	Heading    int32
	RecordedAt time.Time
}

type TrajectoryRepository interface {
	ListByOrder(ctx context.Context, driverID, orderID int64) ([]TrajectoryRecord, error)
	RecordPoint(ctx context.Context, record *TrajectoryRecord) error
}

type HeatmapOrderLocation struct {
	Longitude float64
	Latitude  float64
}

type HeatmapRepository interface {
	ListWaitAcceptOrderLocations(ctx context.Context, longitude, latitude, radiusMeters float64) ([]HeatmapOrderLocation, error)
}

type gormReviewRepository struct {
	db *gorm.DB
}

func NewGormReviewRepository(db *gorm.DB) ReviewRepository {
	return &gormReviewRepository{db: db}
}

func (r *gormReviewRepository) ListByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]PassengerReviewRecord, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&orderReviewRow{}).Where("driver_id = ?", driverID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []orderReviewRow
	offset := int((page - 1) * pageSize)
	if err := query.Order("created_at DESC").Offset(offset).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	records := make([]PassengerReviewRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, PassengerReviewRecord{
			OrderID:   int64(row.OrderID),
			Rating:    int32(row.Rating),
			Comment:   row.Comment,
			CreatedAt: row.CreatedAt,
		})
	}
	return records, total, nil
}

type gormTrajectoryRepository struct {
	db *gorm.DB
}

func NewGormTrajectoryRepository(db *gorm.DB) TrajectoryRepository {
	return &gormTrajectoryRepository{db: db}
}

type gormHeatmapRepository struct {
	db *gorm.DB
}

func NewGormHeatmapRepository(db *gorm.DB) HeatmapRepository {
	return &gormHeatmapRepository{db: db}
}

func (r *gormHeatmapRepository) ListWaitAcceptOrderLocations(ctx context.Context, longitude, latitude, radiusMeters float64) ([]HeatmapOrderLocation, error) {
	minLon, maxLon, minLat, maxLat := heatmapBoundingBox(longitude, latitude, radiusMeters)
	var rows []rideOrderLocationRow
	if err := r.db.WithContext(ctx).
		Select("from_longitude, from_latitude").
		Where("status = ? AND deleted_at IS NULL", constants.OrderStatusWaitAccept).
		Where("from_longitude BETWEEN ? AND ? AND from_latitude BETWEEN ? AND ?", minLon, maxLon, minLat, maxLat).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	locations := make([]HeatmapOrderLocation, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, HeatmapOrderLocation{
			Longitude: row.FromLongitude,
			Latitude:  row.FromLatitude,
		})
	}
	return locations, nil
}

func (r *gormTrajectoryRepository) ListByOrder(ctx context.Context, driverID, orderID int64) ([]TrajectoryRecord, error) {
	var rows []rideTrackPointRow
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND order_id = ?", driverID, orderID).
		Order("recorded_at ASC").
		Limit(maxTrajectoryPoints).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	records := make([]TrajectoryRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, TrajectoryRecord{
			OrderID:    int64(row.OrderID),
			DriverID:   int64(row.DriverID),
			Longitude:  row.Longitude,
			Latitude:   row.Latitude,
			SpeedKmh:   row.SpeedKmh,
			Heading:    int32(row.Direction),
			RecordedAt: row.RecordedAt,
		})
	}
	return records, nil
}

func (r *gormTrajectoryRepository) RecordPoint(ctx context.Context, record *TrajectoryRecord) error {
	if record == nil {
		return nil
	}
	recordedAt := record.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&rideTrackPointRow{
		OrderID:    uint64(record.OrderID),
		DriverID:   uint64(record.DriverID),
		Longitude:  record.Longitude,
		Latitude:   record.Latitude,
		SpeedKmh:   record.SpeedKmh,
		Direction:  int16(record.Heading),
		RecordedAt: recordedAt,
	}).Error
}

type orderReviewRow struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	OrderID   uint64    `gorm:"column:order_id"`
	UserID    uint64    `gorm:"column:user_id"`
	DriverID  uint64    `gorm:"column:driver_id"`
	Rating    int8      `gorm:"column:rating"`
	Comment   string    `gorm:"column:comment"`
	Tags      string    `gorm:"column:tags"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (orderReviewRow) TableName() string {
	return "order_review"
}

type rideTrackPointRow struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	OrderID    uint64    `gorm:"column:order_id"`
	DriverID   uint64    `gorm:"column:driver_id"`
	Longitude  float64   `gorm:"column:longitude"`
	Latitude   float64   `gorm:"column:latitude"`
	SpeedKmh   float64   `gorm:"column:speed_kmh"`
	Direction  int16     `gorm:"column:direction"`
	RecordedAt time.Time `gorm:"column:recorded_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (rideTrackPointRow) TableName() string {
	return "ride_track_point"
}

type rideOrderLocationRow struct {
	FromLongitude float64 `gorm:"column:from_longitude"`
	FromLatitude  float64 `gorm:"column:from_latitude"`
}

func (rideOrderLocationRow) TableName() string {
	return "ride_order"
}

func heatmapBoundingBox(longitude, latitude, radiusMeters float64) (float64, float64, float64, float64) {
	latDelta := radiusMeters / 110540.0
	cosLat := math.Cos(latitude * math.Pi / 180)
	if math.Abs(cosLat) < 0.000001 {
		cosLat = 0.000001
	}
	lonDelta := radiusMeters / (111320.0 * cosLat)
	return longitude - lonDelta, longitude + lonDelta, latitude - latDelta, latitude + latDelta
}

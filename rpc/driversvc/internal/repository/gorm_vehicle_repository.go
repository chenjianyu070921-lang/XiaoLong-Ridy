package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

type gormVehicleRepository struct {
	db *gorm.DB
}

// NewGormVehicleRepository 创建基于 GORM 的车辆仓储。
func NewGormVehicleRepository(db *gorm.DB) DriverVehicleRepository {
	return &gormVehicleRepository{db: db}
}

// Create 写入一条车辆记录。
func (r *gormVehicleRepository) Create(ctx context.Context, vehicle *model.DriverVehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

// GetByID 按主键查询车辆。
func (r *gormVehicleRepository) GetByID(ctx context.Context, id uint64) (*model.DriverVehicle, error) {
	var vehicle model.DriverVehicle
	err := r.db.WithContext(ctx).First(&vehicle, id).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, ErrVehicleNotFound
		}
		return nil, err
	}
	return &vehicle, nil
}

// Update 按 ID 增量更新车辆字段。
func (r *gormVehicleRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.DriverVehicle{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 删除指定车辆。
// 注意：driver_vehicle 表未定义 deleted_at 字段，按物理删除处理。
func (r *gormVehicleRepository) Delete(ctx context.Context, vehicle *model.DriverVehicle) error {
	return r.db.WithContext(ctx).Delete(vehicle).Error
}

package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

// GormDriverHomeDestinationRepository 回家目的地设置的 GORM 实现。
type GormDriverHomeDestinationRepository struct {
	db *gorm.DB
}

// NewGormHomeDestinationRepository 创建回家目的地设置仓储。
func NewGormHomeDestinationRepository(db *gorm.DB) *GormDriverHomeDestinationRepository {
	return &GormDriverHomeDestinationRepository{db: db}
}

// GetByDriver 查询司机回家目的地设置；不存在返回 nil, nil。
func (r *GormDriverHomeDestinationRepository) GetByDriver(ctx context.Context, driverID uint64) (*model.DriverHomeDestination, error) {
	if driverID == 0 {
		return nil, nil
	}
	var setting model.DriverHomeDestination
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Take(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// Upsert 按 driver_id 唯一写入回家目的地设置；已存在则更新地址与模式开关。
func (r *GormDriverHomeDestinationRepository) Upsert(ctx context.Context, setting *model.DriverHomeDestination) error {
	if setting == nil || setting.DriverID == 0 {
		return errors.New("home destination setting required")
	}
	now := time.Now()
	updates := map[string]interface{}{
		"home_addr":         setting.HomeAddr,
		"home_lng":          setting.HomeLng,
		"home_lat":          setting.HomeLat,
		"is_home_mode_open": setting.IsHomeModeOpen,
		"max_detour_ratio":  setting.MaxDetourRatio,
		"update_at":         now,
	}
	result := r.db.WithContext(ctx).Model(&model.DriverHomeDestination{}).
		Where("driver_id = ?", setting.DriverID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	setting.CreateAt = now
	setting.UpdateAt = now
	return r.db.WithContext(ctx).Create(setting).Error
}

// SetModeOpen 切换回家顺路模式开关；未设置过目的地的司机返回 false。
func (r *GormDriverHomeDestinationRepository) SetModeOpen(ctx context.Context, driverID uint64, open bool) (bool, error) {
	if driverID == 0 {
		return false, nil
	}
	openValue := int8(0)
	if open {
		openValue = 1
	}
	result := r.db.WithContext(ctx).Model(&model.DriverHomeDestination{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{"is_home_mode_open": openValue, "update_at": time.Now()})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

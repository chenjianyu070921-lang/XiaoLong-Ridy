package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

// errorsIsNotFound 判断是否为 GORM 记录不存在错误。
func errorsIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

type gormDriverRepository struct {
	db *gorm.DB
}

// NewGormDriverRepository 创建基于 GORM 的司机仓储。
func NewGormDriverRepository(db *gorm.DB) DriverRepository {
	return &gormDriverRepository{db: db}
}

// Create 写入一条司机记录。
func (r *gormDriverRepository) Create(ctx context.Context, driver *model.Driver) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

// GetByID 按主键查询司机，软删记录不可见。
func (r *gormDriverRepository) GetByID(ctx context.Context, id uint64) (*model.Driver, error) {
	var driver model.Driver
	err := r.db.WithContext(ctx).First(&driver, id).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}
	return &driver, nil
}

// Update 按 ID 增量更新司机字段。
func (r *gormDriverRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Driver{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 软删除指定司机（GORM 自动设置 deleted_at）。
func (r *gormDriverRepository) Delete(ctx context.Context, driver *model.Driver) error {
	return r.db.WithContext(ctx).Delete(driver).Error
}

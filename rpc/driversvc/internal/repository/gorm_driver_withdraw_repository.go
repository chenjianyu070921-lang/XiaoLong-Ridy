package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

type gormDriverWithdrawRepository struct {
	db *gorm.DB
}

// NewGormDriverWithdrawRepository 创建基于 GORM 的提现仓储。
func NewGormDriverWithdrawRepository(db *gorm.DB) DriverWithdrawRepository {
	return &gormDriverWithdrawRepository{db: db}
}

// Create 写入一条提现申请记录。
func (r *gormDriverWithdrawRepository) Create(ctx context.Context, withdraw *model.DriverWithdraw) error {
	return r.db.WithContext(ctx).Create(withdraw).Error
}

// ListByDriver 按司机 ID 分页查询提现记录，按申请时间倒序返回本页与总数。
func (r *gormDriverWithdrawRepository) ListByDriver(ctx context.Context, driverID uint64, page, pageSize int32) ([]*model.DriverWithdraw, int64, error) {
	page = clampPage(page)
	pageSize = clampPageSize(pageSize)

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.DriverWithdraw{}).
		Where("driver_id = ?", driverID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []*model.DriverWithdraw
	if err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("applied_at DESC, id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// clampPage 页码收敛：至少 1。
func clampPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

// clampPageSize 每页条数收敛：默认 20，上限 100。
func clampPageSize(pageSize int32) int32 {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

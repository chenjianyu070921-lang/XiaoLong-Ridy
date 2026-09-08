package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

type gormDriverBankCardRepository struct {
	db *gorm.DB
}

// NewGormDriverBankCardRepository 创建基于 GORM 的银行卡仓储。
func NewGormDriverBankCardRepository(db *gorm.DB) DriverBankCardRepository {
	return &gormDriverBankCardRepository{db: db}
}

// Create 写入一张银行卡。
func (r *gormDriverBankCardRepository) Create(ctx context.Context, card *model.DriverBankCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

// ListByDriver 按司机 ID 列出全部未软删银行卡，按 id 升序。
func (r *gormDriverBankCardRepository) ListByDriver(ctx context.Context, driverID uint64) ([]*model.DriverBankCard, error) {
	var cards []*model.DriverBankCard
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND deleted_at IS NULL", driverID).
		Order("id ASC").
		Find(&cards).Error
	return cards, err
}

// CountByDriver 统计司机已绑银行卡数量（不含软删）。
func (r *gormDriverBankCardRepository) CountByDriver(ctx context.Context, driverID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.DriverBankCard{}).
		Where("driver_id = ? AND deleted_at IS NULL", driverID).
		Count(&total).Error
	return total, err
}

// GetByID 按主键查询银行卡（不含软删）。
func (r *gormDriverBankCardRepository) GetByID(ctx context.Context, id uint64) (*model.DriverBankCard, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var card model.DriverBankCard
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&card).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

// Delete 软删除指定银行卡（设置 deleted_at）。
func (r *gormDriverBankCardRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.DriverBankCard{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// UpdateWithdrawPasswordHash 更新司机名下全部银行卡的提现密码哈希（司机级密码冗余同步）。
func (r *gormDriverBankCardRepository) UpdateWithdrawPasswordHash(ctx context.Context, driverID uint64, hash string) error {
	return r.db.WithContext(ctx).
		Model(&model.DriverBankCard{}).
		Where("driver_id = ?", driverID).
		Update("withdraw_password_hash", hash).Error
}

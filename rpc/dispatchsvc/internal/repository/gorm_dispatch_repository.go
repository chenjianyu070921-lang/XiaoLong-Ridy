package repository

import (
	"context"

	"gorm.io/gorm"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
)

type gormDispatchRepository struct {
	db *gorm.DB
}

// NewGormDispatchRepository 创建基于 gorm 的派单记录仓储。
func NewGormDispatchRepository(db *gorm.DB) DispatchRepository {
	return &gormDispatchRepository{db: db}
}

// Create 插入一条派单记录。
func (r *gormDispatchRepository) Create(ctx context.Context, record *model.DispatchRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// ListByOrder 按订单分页查询派单记录，按 ID 正序。
func (r *gormDispatchRepository) ListByOrder(ctx context.Context, orderID uint64, page, pageSize int32) ([]model.DispatchRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).Where("order_id = ?", orderID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.DispatchRecord
	err := q.Order("id ASC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

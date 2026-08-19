package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"XiaoLong-Ridy/common/constants"
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

// RejectByOrderAndDriver 将指定司机对该订单的待派单记录置为已拒绝。
func (r *gormDispatchRepository) RejectByOrderAndDriver(ctx context.Context, orderID, driverID uint64, reason string) (*model.DispatchRecord, error) {
	updates := map[string]interface{}{
		"status": constants.DispatchStatusRejected,
		"remark": reason,
	}
	res := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("order_id = ? AND driver_id = ? AND status = ?", orderID, driverID, constants.DispatchStatusPending).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrDispatchRecordNotFound
	}

	var record model.DispatchRecord
	if err := r.db.WithContext(ctx).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// MarkTimeoutByOrder 将指定订单在 before 之前创建且仍为 Pending 的派单记录置为超时。
func (r *gormDispatchRepository) MarkTimeoutByOrder(ctx context.Context, orderID uint64, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("order_id = ? AND status = ? AND created_at <= ?", orderID, constants.DispatchStatusPending, before).
		Update("status", constants.DispatchStatusTimeout)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// CancelPendingByOrder 将指定订单全部仍为 Pending 的派单记录置为已取消。
func (r *gormDispatchRepository) CancelPendingByOrder(ctx context.Context, orderID uint64, reason string) (int64, error) {
	updates := map[string]interface{}{
		"status": constants.DispatchStatusCancelled,
		"remark": reason,
	}
	res := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("order_id = ? AND status = ?", orderID, constants.DispatchStatusPending).
		Updates(updates)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// ListTimeoutPendingOrderIDs 分页查询存在超时 Pending 派单记录的订单 ID（去重）。
func (r *gormDispatchRepository) ListTimeoutPendingOrderIDs(ctx context.Context, before time.Time, page, pageSize int32) ([]uint64, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("status = ? AND created_at <= ?", constants.DispatchStatusPending, before).
		Distinct("order_id")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ids []uint64
	err := q.Order("order_id ASC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Pluck("order_id", &ids).Error
	if err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}

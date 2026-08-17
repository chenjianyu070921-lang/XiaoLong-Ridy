package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"

	"gorm.io/gorm"
)

// CompleteOrder 条件更新待支付订单为已完成并写入完成日志。
func (r *gormOrderRepository) CompleteOrder(ctx context.Context, orderID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status = ? AND deleted_at IS NULL", orderID, constants.OrderStatusWaitPay).
			Updates(map[string]interface{}{"status": constants.OrderStatusCompleted, "updated_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOrderNotUpdated
		}
		statusLog.OrderId = orderID
		return tx.Create(statusLog).Error
	})
	if errors.Is(err, errOrderNotUpdated) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkDispatchAccepted 将派单中的记录标记为已接受。
func (r *gormOrderRepository) MarkDispatchAccepted(ctx context.Context, orderID, driverID uint64) error {
	res := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("order_id = ? AND driver_id = ? AND status = 1", orderID, driverID).
		Updates(map[string]interface{}{"status": 2, "updated_at": time.Now()})
	return res.Error
}

// CompleteOrder 内存版：待支付订单改为已完成。
func (r *MemoryOrderRepository) CompleteOrder(_ context.Context, orderID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusWaitPay {
		return false, nil
	}

	now := time.Now()
	order.Status = constants.OrderStatusCompleted
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// MarkDispatchAccepted 内存仓储不维护派单记录，测试无需断言。
func (r *MemoryOrderRepository) MarkDispatchAccepted(_ context.Context, _ uint64, _ uint64) error {
	return nil
}

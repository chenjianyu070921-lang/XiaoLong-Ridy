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
// 修复说明（P2-M4-8）：原实现用魔法数字 status=1/2 且无 RowsAffected 校验。
// 改用 DispatchStatusPending/Accepted 常量，并在未命中任何记录时返回明确错误，
// 避免"派单记录已失效却被静默当作成功"导致接单状态不一致。
func (r *gormOrderRepository) MarkDispatchAccepted(ctx context.Context, orderID, driverID uint64) error {
	res := r.db.WithContext(ctx).Model(&model.DispatchRecord{}).
		Where("order_id = ? AND driver_id = ? AND status = ?", orderID, driverID, constants.DispatchStatusPending).
		Updates(map[string]interface{}{"status": constants.DispatchStatusAccepted, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errOrderNotUpdated
	}
	return nil
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

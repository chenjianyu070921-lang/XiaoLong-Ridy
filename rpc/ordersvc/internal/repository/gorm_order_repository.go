package repository

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

var errOrderNotUpdated = errors.New("orderclient not updated")

type gormOrderRepository struct {
	db             *gorm.DB
	couponConsumer CouponConsumer
}

// NewGormOrderRepository 创建基于 gorm 的订单仓储。
func NewGormOrderRepository(db *gorm.DB, couponConsumer CouponConsumer) OrderRepository {
	return &gormOrderRepository{db: db, couponConsumer: couponConsumer}
}

// Create 在事务中创建订单和创建状态日志。
func (r *gormOrderRepository) Create(ctx context.Context, order *model.RideOrder, statusLog *model.OrderStatusLog) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		statusLog.OrderId = order.Id
		return tx.Create(statusLog).Error
	})
	if isDuplicateKey(err) {
		return ErrOrderNoExists
	}
	return err
}

// GetByID 查询未删除的订单完整信息。
func (r *gormOrderRepository) GetByID(ctx context.Context, id uint64) (*model.RideOrder, error) {
	var order model.RideOrder
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Cancel 条件更新订单为已取消，并写入取消日志。
func (r *gormOrderRepository) Cancel(ctx context.Context, orderID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	return r.cancelTx(ctx, r.db, orderID, 0, wantStatuses, cancelBy, reason, statusLog)
}

// CancelWithCoupon 将订单取消和优惠券释放放入同一个 MySQL 事务，任一步失败都会整体回滚。
func (r *gormOrderRepository) CancelWithCoupon(ctx context.Context, orderID, userID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	return r.cancelTx(ctx, r.db, orderID, userID, wantStatuses, cancelBy, reason, statusLog)
}

// cancelTx 承担取消订单的事务细节；userID 大于 0 时同时释放锁定优惠券。
func (r *gormOrderRepository) cancelTx(ctx context.Context, db *gorm.DB, orderID, userID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status IN ? AND deleted_at IS NULL", orderID, wantStatuses).
			Updates(map[string]interface{}{
				"status":        constants.OrderStatusCancelled,
				"cancel_by":     cancelBy,
				"cancel_reason": reason,
				"updated_at":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOrderNotUpdated
		}
		statusLog.OrderId = orderID
		if err := tx.Create(statusLog).Error; err != nil {
			return err
		}
		if userID > 0 {
			if err := tx.Table("user_coupon").Where("user_id = ? AND locked_order_id = ? AND status = ?", userID, orderID, userCouponStatusLocked).
				Updates(map[string]interface{}{"status": userCouponStatusAvailable, "locked_order_id": 0, "locked_at": nil, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, errOrderNotUpdated) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// TimeoutCancel 原子取消仍处于待接单且没有司机接单的订单，避免旧超时任务取消已接单订单。
func (r *gormOrderRepository) TimeoutCancel(ctx context.Context, orderID uint64, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status = ? AND driver_id = 0 AND deleted_at IS NULL", orderID, constants.OrderStatusWaitAccept).
			Updates(map[string]interface{}{
				"status":        constants.OrderStatusCancelled,
				"cancel_by":     constants.OperatorSystem,
				"cancel_reason": reason,
				"updated_at":    time.Now(),
			})
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

// Accept 条件更新待接单订单为已接单并绑定司机。
func (r *gormOrderRepository) Accept(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		dispatchRes := tx.Model(&model.DispatchRecord{}).
			Where("order_id = ? AND driver_id = ? AND status = ?", orderID, driverID, constants.DispatchStatusPending).
			Updates(map[string]interface{}{
				"status":     constants.DispatchStatusAccepted,
				"updated_at": now,
			})
		if dispatchRes.Error != nil {
			return dispatchRes.Error
		}
		if dispatchRes.RowsAffected == 0 {
			return errOrderNotUpdated
		}
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND driver_id = 0 AND status = ? AND deleted_at IS NULL", orderID, constants.OrderStatusWaitAccept).
			Updates(map[string]interface{}{
				"driver_id":  driverID,
				"status":     constants.OrderStatusAccepted,
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOrderNotUpdated
		}
		if err := tx.Model(&model.DispatchRecord{}).
			Where("order_id = ? AND driver_id <> ? AND status = ?", orderID, driverID, constants.DispatchStatusPending).
			Updates(map[string]interface{}{
				"status":     constants.DispatchStatusCancelled,
				"updated_at": now,
			}).Error; err != nil {
			return err
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

// StartTrip 条件更新已接单订单为行程中。
func (r *gormOrderRepository) StartTrip(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND driver_id = ? AND status = ? AND deleted_at IS NULL", orderID, driverID, constants.OrderStatusAccepted).
			Updates(map[string]interface{}{
				"status":     constants.OrderStatusOnTrip,
				"updated_at": time.Now(),
			})
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

// FinishTrip 条件更新行程中订单为待支付。
func (r *gormOrderRepository) FinishTrip(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND driver_id = ? AND status = ? AND deleted_at IS NULL", orderID, driverID, constants.OrderStatusOnTrip).
			Updates(map[string]interface{}{
				"status":     constants.OrderStatusWaitPay,
				"updated_at": time.Now(),
			})
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

// AppendStatusLog 追加一条状态日志。
func (r *gormOrderRepository) AppendStatusLog(ctx context.Context, statusLog *model.OrderStatusLog) error {
	return r.db.WithContext(ctx).Create(statusLog).Error
}

// List 按用户/司机/状态分页查询订单，按 ID 倒序。
func (r *gormOrderRepository) List(ctx context.Context, userID, driverID uint64, status int8, page, pageSize int32) ([]model.RideOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.RideOrder{}).Where("deleted_at IS NULL")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if driverID > 0 {
		q = q.Where("driver_id = ?", driverID)
	}
	if status > 0 {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.RideOrder
	err := q.Order("id DESC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListTimeoutOrders 查询超过超时时间的待接单订单，按创建时间升序。
func (r *gormOrderRepository) ListTimeoutOrders(ctx context.Context, before time.Time, page, pageSize int32) ([]model.RideOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.RideOrder{}).
		Where("status = ? AND created_at <= ? AND deleted_at IS NULL", constants.OrderStatusWaitAccept, before)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.RideOrder
	err := q.Order("created_at ASC, id ASC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListStatusLogs 分页查询订单状态日志，按 ID 正序。
func (r *gormOrderRepository) ListStatusLogs(ctx context.Context, orderID uint64, page, pageSize int32) ([]model.OrderStatusLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.OrderStatusLog{}).Where("order_id = ?", orderID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.OrderStatusLog
	err := q.Order("id ASC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// isDuplicateKey 判断是否为 MySQL 唯一键冲突。
func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// Refund 将已完成订单退款为已退款终态，并累加退款金额。
// 仅允许从已完成状态跳转（由状态机保证），采用 CAS 条件更新防止并发重复退款。
func (r *gormOrderRepository) Refund(ctx context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status = ? AND deleted_at IS NULL", orderID, constants.OrderStatusCompleted).
			Updates(map[string]interface{}{
				"status":       constants.OrderStatusRefunded,
				"refund_cents": gorm.Expr("refund_cents + ?", refundCents),
				"updated_at":   time.Now(),
			})
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

// Redispatch 人工改派：解除司机绑定、订单回到待接单并重新进入派单队列；指定 newDriverID 时直接绑定新司机。
func (r *gormOrderRepository) Redispatch(ctx context.Context, orderID, newDriverID uint64, allowStatuses []int8, statusLog *model.OrderStatusLog) (uint64, bool, error) {
	var finalDriver uint64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将订单从允许的前置状态原子改回待接单并解绑司机。
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status IN ? AND deleted_at IS NULL", orderID, allowStatuses).
			Updates(map[string]interface{}{
				"status":     constants.OrderStatusWaitAccept,
				"driver_id":  0,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOrderNotUpdated
		}
		statusLog.OrderId = orderID
		if err := tx.Create(statusLog).Error; err != nil {
			return err
		}
		// 指定新司机：直接绑定为已接单，跳过自动派单。
		if newDriverID > 0 {
			if err := tx.Model(&model.RideOrder{}).
				Where("id = ? AND status = ? AND deleted_at IS NULL", orderID, constants.OrderStatusWaitAccept).
				Updates(map[string]interface{}{
					"status":     constants.OrderStatusAccepted,
					"driver_id":  newDriverID,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			finalDriver = newDriverID
		}
		return nil
	})
	if errors.Is(err, errOrderNotUpdated) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return finalDriver, true, nil
}

// ForceRefund 管理员强制退款：状态改为已退款并累加退款金额，需状态机合法跳转。
func (r *gormOrderRepository) ForceRefund(ctx context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.RideOrder{}).
			Where("id = ? AND status = ? AND deleted_at IS NULL", orderID, constants.OrderStatusRefunded).
			Updates(map[string]interface{}{
				"refund_cents": gorm.Expr("refund_cents + ?", refundCents),
				"updated_at":   time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		// 若订单已是已退款终态，仅累加退款金额（重复退款场景）。
		if res.RowsAffected == 0 {
			// 否则尝试从其他允许终态跳转至已退款。
			res = tx.Model(&model.RideOrder{}).
				Where("id = ? AND status IN ? AND deleted_at IS NULL", orderID,
					[]int8{constants.OrderStatusCompleted, constants.OrderStatusWaitPay, constants.OrderStatusOnTrip}).
				Updates(map[string]interface{}{
					"status":       constants.OrderStatusRefunded,
					"refund_cents": gorm.Expr("refund_cents + ?", refundCents),
					"updated_at":   time.Now(),
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return errOrderNotUpdated
			}
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

// ReleaseCoupon 释放订单锁定的优惠券（取消/退款时回滚）。
func (r *gormOrderRepository) ReleaseCoupon(ctx context.Context, userID, orderID uint64) error {
	return r.couponConsumer.ReleaseByOrder(ctx, userID, orderID)
}

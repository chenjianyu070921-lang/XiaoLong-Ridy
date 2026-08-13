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

var errOrderNotUpdated = errors.New("order not updated")

type gormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) OrderRepository {
	return &gormOrderRepository{db: db}
}

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

func (r *gormOrderRepository) Cancel(ctx context.Context, orderID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

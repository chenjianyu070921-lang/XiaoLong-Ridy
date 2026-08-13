package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderNoExists = errors.New("order no already exists")
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.RideOrder, statusLog *model.OrderStatusLog) error
	GetByID(ctx context.Context, id uint64) (*model.RideOrder, error)
	Cancel(ctx context.Context, orderID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error)
}

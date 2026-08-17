package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

var (
	// ErrOrderNotFound 表示订单不存在。
	ErrOrderNotFound = errors.New("order not found")
	// ErrOrderNoExists 表示订单号已存在。
	ErrOrderNoExists = errors.New("order no already exists")
)

// OrderRepository 定义订单存储接口。
type OrderRepository interface {
	// Create 创建订单并写入创建状态日志。
	Create(ctx context.Context, order *model.RideOrder, statusLog *model.OrderStatusLog) error
	// GetByID 查询未删除的订单详情。
	GetByID(ctx context.Context, id uint64) (*model.RideOrder, error)
	// Cancel 将订单从允许取消的状态改为已取消，并写入取消日志。
	Cancel(ctx context.Context, orderID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error)
	// TimeoutCancel 仅将未接单且未绑定司机的订单改为超时取消，并写入状态日志。
	TimeoutCancel(ctx context.Context, orderID uint64, reason string, statusLog *model.OrderStatusLog) (bool, error)
	// Accept 将待接单订单改为已接单并绑定司机。
	Accept(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error)
	// StartTrip 将已接单订单改为行程中。
	StartTrip(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error)
	// FinishTrip 将行程中订单改为待支付。
	FinishTrip(ctx context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error)
	// AppendStatusLog 追加一条状态日志。
	AppendStatusLog(ctx context.Context, statusLog *model.OrderStatusLog) error
	// List 按用户/司机/状态分页查询订单。
	List(ctx context.Context, userID, driverID uint64, status int8, page, pageSize int32) ([]model.RideOrder, int64, error)
	// ListTimeoutOrders 查询创建时间早于 before 的待接单订单。
	ListTimeoutOrders(ctx context.Context, before time.Time, page, pageSize int32) ([]model.RideOrder, int64, error)
	// ListStatusLogs 分页查询订单状态日志。
	ListStatusLogs(ctx context.Context, orderID uint64, page, pageSize int32) ([]model.OrderStatusLog, int64, error)
}

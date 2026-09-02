package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

var (
	// ErrOrderNotFound 表示订单不存在。
	ErrOrderNotFound = errors.New("orderclient not found")
	// ErrOrderNoExists 表示订单号已存在。
	ErrOrderNoExists = errors.New("orderclient no already exists")
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
	// CompleteOrder 将待支付订单改为已完成，写入完成日志，并落库实付金额 paidCents。
	// paidCents 必落：下游退款金额取 order.PaidCents，若不落库则恒为 0，退款链路完全失效。
	CompleteOrder(ctx context.Context, orderID uint64, statusLog *model.OrderStatusLog, paidCents int64) (bool, error)
	// MarkDispatchAccepted 将指定司机的派单记录标记为已接受。
	MarkDispatchAccepted(ctx context.Context, orderID, driverID uint64) error
	// AppendStatusLog 追加一条状态日志。
	AppendStatusLog(ctx context.Context, statusLog *model.OrderStatusLog) error
	// List 按用户/司机/状态分页查询订单。
	List(ctx context.Context, userID, driverID uint64, status int8, page, pageSize int32) ([]model.RideOrder, int64, error)
	// ListTimeoutOrders 查询创建时间早于 before 的待接单订单。
	ListTimeoutOrders(ctx context.Context, before time.Time, page, pageSize int32) ([]model.RideOrder, int64, error)
	// ListStatusLogs 分页查询订单状态日志。
	ListStatusLogs(ctx context.Context, orderID uint64, page, pageSize int32) ([]model.OrderStatusLog, int64, error)
	// Refund 将已完成订单退款为已退款终态，并落库退款金额，需状态机合法跳转。
	Refund(ctx context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error)
	// ReleaseCoupon 释放订单锁定的优惠券（取消/退款时回滚），幂等。
	ReleaseCoupon(ctx context.Context, userID, orderID uint64) error
	// Redispatch 人工改派：解除司机绑定、订单回到待接单并重新进入派单队列；指定 newDriverID 时直接绑定新司机。
	// allowStatuses 为该操作允许的前置状态（如待接单/已接单）。返回最终绑定的 driverID（0 表示回自动派单池）。
	Redispatch(ctx context.Context, orderID, newDriverID uint64, allowStatuses []int8, statusLog *model.OrderStatusLog) (uint64, bool, error)
	// ForceRefund 管理员强制退款：可从更多终态（含已支付/行程中等）发起，状态改为已退款并累加退款金额，需状态机合法跳转。
	ForceRefund(ctx context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error)
}

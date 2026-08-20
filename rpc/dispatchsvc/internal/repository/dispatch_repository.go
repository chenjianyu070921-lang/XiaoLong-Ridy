package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
)

var (
	// ErrDispatchRecordNotFound 表示派单记录不存在。
	ErrDispatchRecordNotFound = errors.New("dispatch record not found")
)

// DispatchRepository 定义派单记录存储接口。
type DispatchRepository interface {
	Create(ctx context.Context, record *model.DispatchRecord) error
	ListByOrder(ctx context.Context, orderID uint64, page, pageSize int32) ([]model.DispatchRecord, int64, error)
	// ListByDriver 按司机分页查询派单记录；status > 0 时按派单状态过滤，0 表示全部。按 ID 正序。
	ListByDriver(ctx context.Context, driverID uint64, status int32, page, pageSize int32) ([]model.DispatchRecord, int64, error)
	// RejectByOrderAndDriver 将指定司机对该订单的待派单记录置为已拒绝。
	// 当该司机不存在待派单记录时返回 ErrDispatchRecordNotFound。
	RejectByOrderAndDriver(ctx context.Context, orderID, driverID uint64, reason string) (*model.DispatchRecord, error)
	// MarkTimeoutByOrder 将指定订单在 before 之前创建且仍为 Pending 的派单记录置为超时，
	// 返回受影响记录数。
	MarkTimeoutByOrder(ctx context.Context, orderID uint64, before time.Time) (int64, error)
	// CancelPendingByOrder 将指定订单全部仍为 Pending 的派单记录置为已取消，
	// 返回受影响记录数。
	CancelPendingByOrder(ctx context.Context, orderID uint64, reason string) (int64, error)
	// ListTimeoutPendingOrderIDs 分页查询存在超时（created_at <= before 且仍为 Pending）
	// 派单记录的订单 ID（去重），返回 ID 列表与订单总数。
	ListTimeoutPendingOrderIDs(ctx context.Context, before time.Time, page, pageSize int32) ([]uint64, int64, error)
}

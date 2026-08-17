package repository

import (
	"context"
	"errors"

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
}

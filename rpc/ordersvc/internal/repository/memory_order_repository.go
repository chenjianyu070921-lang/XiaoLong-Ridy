package repository

import (
	"context"
	"sync"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

type MemoryOrderRepository struct {
	mu        sync.RWMutex
	nextID    uint64
	nextLogID uint64
	orders    map[uint64]*model.RideOrder
	orderLogs map[uint64][]model.OrderStatusLog
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		nextID:    1,
		nextLogID: 1,
		orders:    make(map[uint64]*model.RideOrder),
		orderLogs: make(map[uint64][]model.OrderStatusLog),
	}
}

func (r *MemoryOrderRepository) Create(_ context.Context, order *model.RideOrder, statusLog *model.OrderStatusLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.orders {
		if existing.OrderNo == order.OrderNo {
			return ErrOrderNoExists
		}
	}

	now := time.Now()
	copied := *order
	copied.Id = r.nextID
	copied.CreatedAt = now
	copied.UpdatedAt = now

	logCopied := *statusLog
	logCopied.Id = r.nextLogID
	logCopied.OrderId = copied.Id
	logCopied.CreatedAt = now

	r.nextID++
	r.nextLogID++
	r.orders[copied.Id] = &copied
	r.orderLogs[copied.Id] = append(r.orderLogs[copied.Id], logCopied)

	order.Id = copied.Id
	order.CreatedAt = copied.CreatedAt
	order.UpdatedAt = copied.UpdatedAt
	statusLog.Id = logCopied.Id
	statusLog.OrderId = logCopied.OrderId
	statusLog.CreatedAt = logCopied.CreatedAt
	return nil
}

func (r *MemoryOrderRepository) GetByID(_ context.Context, id uint64) (*model.RideOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[id]
	if !ok {
		return nil, ErrOrderNotFound
	}
	copied := *order
	return &copied, nil
}

func (r *MemoryOrderRepository) Cancel(_ context.Context, orderID uint64, wantStatuses []int8, cancelBy, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if !containsStatus(wantStatuses, order.Status) {
		return false, nil
	}

	now := time.Now()
	order.Status = constants.OrderStatusCancelled
	order.CancelBy = cancelBy
	order.CancelReason = reason
	order.UpdatedAt = now

	logCopied := *statusLog
	logCopied.Id = r.nextLogID
	logCopied.OrderId = orderID
	logCopied.CreatedAt = now
	r.nextLogID++
	r.orderLogs[orderID] = append(r.orderLogs[orderID], logCopied)

	statusLog.Id = logCopied.Id
	statusLog.OrderId = logCopied.OrderId
	statusLog.CreatedAt = logCopied.CreatedAt
	return true, nil
}

func (r *MemoryOrderRepository) StatusLogs(orderID uint64) []model.OrderStatusLog {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logs := make([]model.OrderStatusLog, len(r.orderLogs[orderID]))
	copy(logs, r.orderLogs[orderID])
	return logs
}

func containsStatus(statuses []int8, want int8) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}
	return false
}

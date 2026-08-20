package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
)

// MemoryDispatchRepository 是用于本地开发与测试的内存派单记录仓储。
type MemoryDispatchRepository struct {
	mu      sync.RWMutex
	nextID  uint64
	records map[uint64]model.DispatchRecord
}

// NewMemoryDispatchRepository 创建内存派单记录仓储。
func NewMemoryDispatchRepository() *MemoryDispatchRepository {
	return &MemoryDispatchRepository{
		nextID:  1,
		records: make(map[uint64]model.DispatchRecord),
	}
}

// Create 插入一条派单记录并回填 ID 与时间。
func (r *MemoryDispatchRepository) Create(_ context.Context, record *model.DispatchRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	copied := *record
	copied.Id = r.nextID
	copied.CreatedAt = now
	copied.UpdatedAt = now
	r.nextID++
	r.records[copied.Id] = copied
	*record = copied
	return nil
}

// ListByOrder 按订单分页查询派单记录，按 ID 正序。
func (r *MemoryDispatchRepository) ListByOrder(_ context.Context, orderID uint64, page, pageSize int32) ([]model.DispatchRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uint64, 0)
	for id, record := range r.records {
		if record.OrderId == orderID {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	total := int64(len(ids))
	start := int((page - 1) * pageSize)
	if start >= len(ids) {
		return []model.DispatchRecord{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(ids) {
		end = len(ids)
	}
	out := make([]model.DispatchRecord, 0, end-start)
	for _, id := range ids[start:end] {
		out = append(out, r.records[id])
	}
	return out, total, nil
}

// ListByDriver 按司机分页查询派单记录，status > 0 时过滤，按 ID 正序。
func (r *MemoryDispatchRepository) ListByDriver(_ context.Context, driverID uint64, status int32, page, pageSize int32) ([]model.DispatchRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uint64, 0)
	for id, record := range r.records {
		if record.DriverId == driverID && (status == 0 || int32(record.Status) == status) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	total := int64(len(ids))
	start := int((page - 1) * pageSize)
	if start >= len(ids) {
		return []model.DispatchRecord{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(ids) {
		end = len(ids)
	}
	out := make([]model.DispatchRecord, 0, end-start)
	for _, id := range ids[start:end] {
		out = append(out, r.records[id])
	}
	return out, total, nil
}

// RejectByOrderAndDriver 将指定司机对该订单的待派单记录置为已拒绝。
func (r *MemoryDispatchRepository) RejectByOrderAndDriver(_ context.Context, orderID, driverID uint64, reason string) (*model.DispatchRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var target *model.DispatchRecord
	for id, record := range r.records {
		if record.OrderId == orderID && record.DriverId == driverID && record.Status == constants.DispatchStatusPending {
			copied := record
			copied.Status = constants.DispatchStatusRejected
			copied.Remark = reason
			copied.UpdatedAt = time.Now()
			r.records[id] = copied
			target = &copied
			break
		}
	}
	if target == nil {
		return nil, ErrDispatchRecordNotFound
	}
	return target, nil
}

// MarkTimeoutByOrder 将指定订单在 before 之前创建且仍为 Pending 的派单记录置为超时。
func (r *MemoryDispatchRepository) MarkTimeoutByOrder(_ context.Context, orderID uint64, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var affected int64
	for id, record := range r.records {
		if record.OrderId == orderID && record.Status == constants.DispatchStatusPending && !record.CreatedAt.After(before) {
			copied := record
			copied.Status = constants.DispatchStatusTimeout
			copied.UpdatedAt = time.Now()
			r.records[id] = copied
			affected++
		}
	}
	return affected, nil
}

// CancelPendingByOrder 将指定订单全部仍为 Pending 的派单记录置为已取消。
func (r *MemoryDispatchRepository) CancelPendingByOrder(_ context.Context, orderID uint64, reason string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var affected int64
	for id, record := range r.records {
		if record.OrderId == orderID && record.Status == constants.DispatchStatusPending {
			copied := record
			copied.Status = constants.DispatchStatusCancelled
			copied.Remark = reason
			copied.UpdatedAt = time.Now()
			r.records[id] = copied
			affected++
		}
	}
	return affected, nil
}

// ListTimeoutPendingOrderIDs 分页查询存在超时 Pending 派单记录的订单 ID（去重）。
func (r *MemoryDispatchRepository) ListTimeoutPendingOrderIDs(_ context.Context, before time.Time, page, pageSize int32) ([]uint64, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orderSet := make(map[uint64]struct{})
	for _, record := range r.records {
		if record.Status == constants.DispatchStatusPending && !record.CreatedAt.After(before) {
			orderSet[record.OrderId] = struct{}{}
		}
	}
	ids := make([]uint64, 0, len(orderSet))
	for id := range orderSet {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	total := int64(len(ids))
	start := int((page - 1) * pageSize)
	if start >= len(ids) {
		return []uint64{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(ids) {
		end = len(ids)
	}
	return ids[start:end], total, nil
}

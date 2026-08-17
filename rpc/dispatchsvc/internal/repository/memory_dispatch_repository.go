package repository

import (
	"context"
	"sort"
	"sync"
	"time"

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

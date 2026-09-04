package repository

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
)

func TestMemoryDispatchRepositoryStatusFlow(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryDispatchRepository()

	old := time.Now().Add(-5 * time.Minute)
	now := time.Now()

	// 订单 1：两条超时 Pending + 一条新建 Pending（超时阈值判定）。
	for _, driver := range []uint64{9001, 9002, 9003} {
		rec := &model.DispatchRecord{OrderId: 1, DriverId: driver, Status: constants.DispatchStatusPending}
		if err := repo.Create(ctx, rec); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		// 回拨创建时间，模拟超时。
		repo.mu.Lock()
		if driver != 9003 {
			copied := repo.records[rec.Id]
			copied.CreatedAt = old
			repo.records[rec.Id] = copied
		}
		repo.mu.Unlock()
		_ = now
	}

	// 订单 2：已取消，不应出现。
	cancelled := &model.DispatchRecord{OrderId: 2, DriverId: 9101, Status: constants.DispatchStatusCancelled, CreatedAt: old, UpdatedAt: now}
	if err := repo.Create(ctx, cancelled); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	repo.mu.Lock()
	copied := repo.records[cancelled.Id]
	copied.CreatedAt = old
	repo.records[cancelled.Id] = copied
	repo.mu.Unlock()

	// 超时订单列表：只有订单 1（两条超时 Pending 去重为 1 个订单）。
	ids, total, err := repo.ListTimeoutPendingOrderIDs(ctx, time.Now().Add(-time.Minute), 1, 10)
	if err != nil {
		t.Fatalf("ListTimeoutPendingOrderIDs() error = %v", err)
	}
	if total != 1 || len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("ListTimeoutPendingOrderIDs() ids=%v total=%d, want [1]/1", ids, total)
	}

	// MarkTimeoutByOrder：订单 1 在 before=now 之前创建且 Pending 的记录全部置超时。
	affected, err := repo.MarkTimeoutByOrder(ctx, 1, time.Now())
	if err != nil {
		t.Fatalf("MarkTimeoutByOrder() error = %v", err)
	}
	if affected != 3 {
		t.Fatalf("MarkTimeoutByOrder() affected = %d, want 3", affected)
	}
	// 全部置超时后，超时扫描不再返回订单 1（已非 Pending）。
	if _, total, err := repo.ListTimeoutPendingOrderIDs(ctx, time.Now(), 1, 10); err != nil || total != 0 {
		t.Fatalf("ListTimeoutPendingOrderIDs() after timeout total=%d err=%v, want 0", total, err)
	}

	// 重新派单（新 Pending 记录）后超时列表应再次可见（人工回拨时间）。
	fresh := &model.DispatchRecord{OrderId: 1, DriverId: 9004, Status: constants.DispatchStatusPending}
	if err := repo.Create(ctx, fresh); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	repo.mu.Lock()
	copied = repo.records[fresh.Id]
	copied.CreatedAt = old
	repo.records[fresh.Id] = copied
	repo.mu.Unlock()
	ids, total, _ = repo.ListTimeoutPendingOrderIDs(ctx, time.Now().Add(-time.Minute), 1, 10)
	if total != 1 || ids[0] != 1 {
		t.Fatalf("ListTimeoutPendingOrderIDs() after fresh pending = %v/%d, want [1]/1", ids, total)
	}

	// CancelPendingByOrder：取消订单 1 全部 Pending。
	affected, err = repo.CancelPendingByOrder(ctx, 1, "用户取消")
	if err != nil {
		t.Fatalf("CancelPendingByOrder() error = %v", err)
	}
	if affected != 1 {
		t.Fatalf("CancelPendingByOrder() affected = %d, want 1", affected)
	}
	if _, total, err := repo.ListTimeoutPendingOrderIDs(ctx, time.Now(), 1, 10); err != nil || total != 0 {
		t.Fatalf("ListTimeoutPendingOrderIDs() after cancel total=%d err=%v, want 0", total, err)
	}
}

// TestMemoryDispatchRepositoryPagination 超时订单分页。
func TestMemoryDispatchRepositoryPagination(t *testing.T) {
	repo := NewMemoryDispatchRepository()
	old := time.Now().Add(-10 * time.Minute)
	for i := uint64(1); i <= 3; i++ {
		rec := &model.DispatchRecord{OrderId: i, DriverId: 1000 + i, Status: constants.DispatchStatusPending}
		if err := repo.Create(context.Background(), rec); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		repo.mu.Lock()
		copied := repo.records[rec.Id]
		copied.CreatedAt = old
		repo.records[rec.Id] = copied
		repo.mu.Unlock()
	}

	ids, total, err := repo.ListTimeoutPendingOrderIDs(context.Background(), time.Now(), 2, 2)
	if err != nil {
		t.Fatalf("ListTimeoutPendingOrderIDs() error = %v", err)
	}
	if total != 3 || len(ids) != 1 || ids[0] != 3 {
		t.Fatalf("ListTimeoutPendingOrderIDs() page2 ids=%v total=%d, want [3]/3", ids, total)
	}
}

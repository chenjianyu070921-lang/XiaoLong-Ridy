package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
)

type MemoryOrderRepository struct {
	mu              sync.RWMutex
	nextID          uint64
	nextLogID       uint64
	nextDispatchID  uint64
	orders          map[uint64]*model.RideOrder
	orderLogs       map[uint64][]model.OrderStatusLog
	dispatchRecords map[uint64]*model.DispatchRecord
}

// NewMemoryOrderRepository 创建用于本地开发与测试的内存订单仓储。
func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		nextID:          1,
		nextLogID:       1,
		nextDispatchID:  1,
		orders:          make(map[uint64]*model.RideOrder),
		orderLogs:       make(map[uint64][]model.OrderStatusLog),
		dispatchRecords: make(map[uint64]*model.DispatchRecord),
	}
}

// Create 在内存中创建订单并追加创建状态日志。
func (r *MemoryOrderRepository) Create(_ context.Context, order *model.RideOrder, statusLog *model.OrderStatusLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.orders {
		if existing.OrderNo == order.OrderNo {
			return ErrOrderNoExists
		}
	}

	now := time.Now()
	createdAt := order.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	copied := *order
	copied.Id = r.nextID
	copied.CreatedAt = createdAt
	copied.UpdatedAt = now

	logCopied := *statusLog
	logCopied.Id = r.nextLogID
	logCopied.OrderId = copied.Id
	logCopied.CreatedAt = createdAt

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

// GetByID 返回订单副本。
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

// Cancel 条件更新订单为已取消并追加日志。
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
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// TimeoutCancel 只允许超时任务取消仍处于待接单且未绑定司机的订单。
func (r *MemoryOrderRepository) TimeoutCancel(_ context.Context, orderID uint64, reason string, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusWaitAccept || order.DriverId != 0 {
		return false, nil
	}

	now := time.Now()
	order.Status = constants.OrderStatusCancelled
	order.CancelBy = constants.OperatorSystem
	order.CancelReason = reason
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// Accept 将待接单订单改为已接单并绑定司机。
func (r *MemoryOrderRepository) Accept(_ context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusWaitAccept || order.DriverId != 0 {
		return false, nil
	}
	dispatchRecord := r.findPendingDispatchRecordLocked(orderID, driverID)
	if dispatchRecord == nil {
		return false, nil
	}

	now := time.Now()
	dispatchRecord.Status = constants.DispatchStatusAccepted
	dispatchRecord.UpdatedAt = now
	r.cancelOtherPendingDispatchRecordsLocked(orderID, driverID, now)
	order.Status = constants.OrderStatusAccepted
	order.DriverId = driverID
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// StartTrip 将已接单订单改为行程中。
func (r *MemoryOrderRepository) StartTrip(_ context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusAccepted || order.DriverId != driverID {
		return false, nil
	}

	now := time.Now()
	order.Status = constants.OrderStatusOnTrip
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// FinishTrip 将行程中订单改为待支付。
func (r *MemoryOrderRepository) FinishTrip(_ context.Context, orderID, driverID uint64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusOnTrip || order.DriverId != driverID {
		return false, nil
	}

	now := time.Now()
	order.Status = constants.OrderStatusWaitPay
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// AppendStatusLog 追加一条状态日志。
func (r *MemoryOrderRepository) AppendStatusLog(_ context.Context, statusLog *model.OrderStatusLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.appendLogLocked(statusLog.OrderId, statusLog, time.Now())
	return nil
}

// List 按用户/司机/状态分页查询订单，按 ID 倒序。
func (r *MemoryOrderRepository) List(_ context.Context, userID, driverID uint64, status int8, page, pageSize int32) ([]model.RideOrder, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uint64, 0, len(r.orders))
	for id := range r.orders {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })

	filtered := make([]model.RideOrder, 0, len(ids))
	for _, id := range ids {
		order := r.orders[id]
		if (userID == 0 || order.UserId == userID) &&
			(driverID == 0 || order.DriverId == driverID) &&
			(status == 0 || order.Status == status) {
			filtered = append(filtered, *order)
		}
	}

	total := int64(len(filtered))
	start := int((page - 1) * pageSize)
	if start >= len(filtered) {
		return []model.RideOrder{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

// ListTimeoutOrders 查询超过超时时间的待接单订单，按创建时间升序。
func (r *MemoryOrderRepository) ListTimeoutOrders(_ context.Context, before time.Time, page, pageSize int32) ([]model.RideOrder, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uint64, 0, len(r.orders))
	for id, order := range r.orders {
		if order.Status == constants.OrderStatusWaitAccept && !order.CreatedAt.After(before) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		left := r.orders[ids[i]]
		right := r.orders[ids[j]]
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.Id < right.Id
		}
		return left.CreatedAt.Before(right.CreatedAt)
	})

	total := int64(len(ids))
	start := int((page - 1) * pageSize)
	if start >= len(ids) {
		return []model.RideOrder{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(ids) {
		end = len(ids)
	}
	out := make([]model.RideOrder, 0, end-start)
	for _, id := range ids[start:end] {
		out = append(out, *r.orders[id])
	}
	return out, total, nil
}

// ListStatusLogs 分页查询订单状态日志，按写入顺序返回。
func (r *MemoryOrderRepository) ListStatusLogs(_ context.Context, orderID uint64, page, pageSize int32) ([]model.OrderStatusLog, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logs := r.orderLogs[orderID]
	total := int64(len(logs))
	start := int((page - 1) * pageSize)
	if start >= len(logs) {
		return []model.OrderStatusLog{}, total, nil
	}
	end := start + int(pageSize)
	if end > len(logs) {
		end = len(logs)
	}
	out := make([]model.OrderStatusLog, 0, end-start)
	out = append(out, logs[start:end]...)
	return out, total, nil
}

// Refund 内存版：已完成订单退款为已退款终态并累加退款金额。
func (r *MemoryOrderRepository) Refund(_ context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status != constants.OrderStatusCompleted {
		return false, nil
	}
	now := time.Now()
	order.Status = constants.OrderStatusRefunded
	order.RefundCents += refundCents
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// Redispatch 人工改派：解除司机绑定、订单回到待接单；指定 newDriverID 时直接绑定新司机。
func (r *MemoryOrderRepository) Redispatch(_ context.Context, orderID, newDriverID uint64, allowStatuses []int8, statusLog *model.OrderStatusLog) (uint64, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return 0, false, ErrOrderNotFound
	}
	if !containsStatus(allowStatuses, order.Status) {
		return 0, false, nil
	}
	now := time.Now()
	order.Status = constants.OrderStatusWaitAccept
	order.DriverId = 0
	order.UpdatedAt = now
	var finalDriver uint64
	if newDriverID > 0 {
		order.Status = constants.OrderStatusAccepted
		order.DriverId = newDriverID
		finalDriver = newDriverID
	}
	r.appendLogLocked(orderID, statusLog, now)
	return finalDriver, true, nil
}

// ForceRefund 管理员强制退款：从允许终态或已退款态累加退款金额。
func (r *MemoryOrderRepository) ForceRefund(_ context.Context, orderID uint64, refundCents int64, statusLog *model.OrderStatusLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, ok := r.orders[orderID]
	if !ok {
		return false, ErrOrderNotFound
	}
	if order.Status == constants.OrderStatusRefunded {
		// 已退款：仅累加金额（重复退款场景）。
		now := time.Now()
		order.RefundCents += refundCents
		order.UpdatedAt = now
		r.appendLogLocked(orderID, statusLog, now)
		return true, nil
	}
	allowed := []int8{constants.OrderStatusCompleted, constants.OrderStatusWaitPay, constants.OrderStatusOnTrip}
	if !containsStatus(allowed, order.Status) {
		return false, nil
	}
	now := time.Now()
	order.Status = constants.OrderStatusRefunded
	order.RefundCents += refundCents
	order.UpdatedAt = now
	r.appendLogLocked(orderID, statusLog, now)
	return true, nil
}

// ReleaseCoupon 内存版：仅记录调用，不做真实释放。
func (r *MemoryOrderRepository) ReleaseCoupon(_ context.Context, _ uint64, _ uint64) error {
	return nil
}

// appendLogLocked 在持有写锁时追加日志并分配 ID 和时间。
func (r *MemoryOrderRepository) appendLogLocked(orderID uint64, statusLog *model.OrderStatusLog, now time.Time) {
	logCopied := *statusLog
	logCopied.Id = r.nextLogID
	logCopied.OrderId = orderID
	logCopied.CreatedAt = now
	r.nextLogID++
	r.orderLogs[orderID] = append(r.orderLogs[orderID], logCopied)

	statusLog.Id = logCopied.Id
	statusLog.OrderId = logCopied.OrderId
	statusLog.CreatedAt = logCopied.CreatedAt
}

// StatusLogs 返回订单全部状态日志，供测试断言使用。
func (r *MemoryOrderRepository) StatusLogs(orderID uint64) []model.OrderStatusLog {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logs := make([]model.OrderStatusLog, len(r.orderLogs[orderID]))
	copy(logs, r.orderLogs[orderID])
	return logs
}

// CreateDispatchRecord 写入派单记录，供本地联调和测试模拟 dispatchsvc 已生成候选司机。
func (r *MemoryOrderRepository) CreateDispatchRecord(_ context.Context, record *model.DispatchRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	copied := *record
	copied.Id = r.nextDispatchID
	copied.CreatedAt = now
	copied.UpdatedAt = now
	r.nextDispatchID++
	r.dispatchRecords[copied.Id] = &copied
	*record = copied
	return nil
}

// DispatchRecords 返回指定订单派单记录副本，供测试断言 claim 结果。
func (r *MemoryOrderRepository) DispatchRecords(orderID uint64) []model.DispatchRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records := make([]model.DispatchRecord, 0)
	for _, record := range r.dispatchRecords {
		if record.OrderId == orderID {
			records = append(records, *record)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Id < records[j].Id })
	return records
}

func (r *MemoryOrderRepository) findPendingDispatchRecordLocked(orderID, driverID uint64) *model.DispatchRecord {
	for _, record := range r.dispatchRecords {
		if record.OrderId == orderID && record.DriverId == driverID && record.Status == constants.DispatchStatusPending {
			return record
		}
	}
	return nil
}

func (r *MemoryOrderRepository) cancelOtherPendingDispatchRecordsLocked(orderID, acceptedDriverID uint64, now time.Time) {
	for _, record := range r.dispatchRecords {
		if record.OrderId == orderID && record.DriverId != acceptedDriverID && record.Status == constants.DispatchStatusPending {
			record.Status = constants.DispatchStatusCancelled
			record.UpdatedAt = now
		}
	}
}

// containsStatus 判断状态列表是否包含指定状态。
func containsStatus(statuses []int8, want int8) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}
	return false
}

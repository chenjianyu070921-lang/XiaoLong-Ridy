package task

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/orderclient"
	"gorm.io/gorm"
)

// RefundCompensationTask 对持久化退款补偿任务执行抢占、重试和终态治理。
type RefundCompensationTask struct {
	db     *gorm.DB
	orders orderclient.Order
}

// NewRefundCompensationTask 创建退款补偿任务处理器。
func NewRefundCompensationTask(db *gorm.DB, orders orderclient.Order) *RefundCompensationTask {
	return &RefundCompensationTask{db: db, orders: orders}
}

// RunRefundCompensation 扫描到期任务并执行有限批量退款。
func (t *RefundCompensationTask) RunRefundCompensation(max int) error {
	if t == nil || t.db == nil || t.orders == nil {
		return fmt.Errorf("退款补偿依赖未配置")
	}
	if max <= 0 {
		max = 50
	}
	var rows []struct {
		ID          int64
		OrderID     int64
		RefundNo    string
		RefundCents int64
		Reason      string
		RetryCount  int
		MaxRetry    int
	}
	now := time.Now()
	if err := t.db.Table("admin_refund_compensation_task").Where("status IN ('pending','retrying') AND (next_retry_at IS NULL OR next_retry_at<=?)", now).Order("id ASC").Limit(max).Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		owner := fmt.Sprintf("refund-%d", time.Now().UnixNano())
		claim := t.db.Exec("UPDATE admin_refund_compensation_task SET status='processing', handled_by=0, updated_at=NOW() WHERE id=? AND status IN ('pending','retrying')", row.ID)
		if claim.Error != nil || claim.RowsAffected == 0 {
			continue
		}
		_, err := t.orders.ForceRefundOrder(context.Background(), &orderclient.ForceRefundOrderRequest{
			OrderId: row.OrderID, RefundNo: row.RefundNo, RefundAmountCents: row.RefundCents, Reason: row.Reason,
		})
		if err == nil {
			_ = t.db.Exec("UPDATE admin_refund_compensation_task SET status='success', failure_reason='', last_response=?, updated_at=NOW() WHERE id=? AND status='processing'", owner, row.ID).Error
			continue
		}
		nextRetry := row.RetryCount + 1
		if nextRetry >= row.MaxRetry {
			_ = t.db.Exec("UPDATE admin_refund_compensation_task SET status='manual_review', retry_count=?, failure_reason=?, updated_at=NOW() WHERE id=? AND status='processing'", nextRetry, err.Error(), row.ID).Error
			continue
		}
		delay := time.Duration(1<<min(nextRetry, 6)) * time.Minute
		_ = t.db.Exec("UPDATE admin_refund_compensation_task SET status='retrying', retry_count=?, next_retry_at=?, failure_reason=?, updated_at=NOW() WHERE id=? AND status='processing'", nextRetry, time.Now().Add(delay), err.Error(), row.ID).Error
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

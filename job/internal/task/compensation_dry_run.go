package task

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/common/constants"
	"github.com/redis/go-redis/v9"
)

// CompensationSummary 是补偿队列只读检查结果，不包含任何业务载荷。
type CompensationSummary struct {
	GeneratedAt      time.Time     `json:"generated_at"`
	RefundEvents     QueueSummary  `json:"refund_events"`
	DispatchRetries  QueueSummary  `json:"dispatch_retries"`
	AdminAuditOutbox OutboxSummary `json:"admin_audit_outbox"`
}

// readQueueSummary 只读取 Redis ZSet 的数量和最早到期时间。
func readQueueSummary(ctx context.Context, client *redis.Client, key string) (QueueSummary, error) {
	result := QueueSummary{Key: key}
	count, err := client.ZCard(ctx, key).Result()
	if err != nil {
		return result, err
	}
	result.Pending = count
	items, err := client.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		return result, err
	}
	if len(items) > 0 {
		result.EarliestAt = time.Unix(int64(items[0].Score), 0)
	}
	return result, nil
}

// QueueSummary 表示 Redis 延迟队列的积压概况。
type QueueSummary struct {
	Key        string    `json:"key"`
	Pending    int64     `json:"pending"`
	EarliestAt time.Time `json:"earliest_at,omitempty"`
}

// OutboxSummary 表示管理后台审计补偿表的只读概况。
type OutboxSummary struct {
	Pending    int64     `json:"pending"`
	Running    int64     `json:"running"`
	Failed     int64     `json:"failed"`
	EarliestAt time.Time `json:"earliest_at,omitempty"`
}

// DryRunCompensationSummary 只读取 Redis 队列和 admin_audit_outbox，供部署前验证依赖和积压。
// 该方法不会发送消息、调用下游 RPC、更新 Redis 或写入数据库。
func (t *Task) DryRunCompensationSummary(ctx context.Context) (*CompensationSummary, error) {
	if t == nil || t.svcCtx == nil {
		return nil, errors.New("compensation dry-run requires service context")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	result := &CompensationSummary{GeneratedAt: time.Now()}
	if t.svcCtx.Redis == nil {
		return nil, errors.New("compensation dry-run requires redis")
	}
	var err error
	if result.RefundEvents, err = readQueueSummary(ctx, t.svcCtx.Redis, constants.RefundRetryQueueKey); err != nil {
		return nil, err
	}
	if result.DispatchRetries, err = readQueueSummary(ctx, t.svcCtx.Redis, constants.DispatchRetryQueueKey); err != nil {
		return nil, err
	}
	if t.svcCtx.Db == nil {
		return nil, errors.New("compensation dry-run requires mysql")
	}
	var pending, running, failed int64
	if err := t.svcCtx.Db.WithContext(ctx).Table("admin_audit_outbox").Where("status = ?", adminOutboxStatusPending).Count(&pending).Error; err != nil {
		return nil, err
	}
	if err := t.svcCtx.Db.WithContext(ctx).Table("admin_audit_outbox").Where("status = ?", adminOutboxStatusRunning).Count(&running).Error; err != nil {
		return nil, err
	}
	if err := t.svcCtx.Db.WithContext(ctx).Table("admin_audit_outbox").Where("status = ?", adminOutboxStatusFailed).Count(&failed).Error; err != nil {
		return nil, err
	}
	result.AdminAuditOutbox = OutboxSummary{Pending: pending, Running: running, Failed: failed}
	var earliest time.Time
	if err := t.svcCtx.Db.WithContext(ctx).Table("admin_audit_outbox").Where("status IN ?", []string{adminOutboxStatusPending, adminOutboxStatusRunning}).Select("MIN(created_at)").Scan(&earliest).Error; err != nil {
		return nil, err
	}
	result.AdminAuditOutbox.EarliestAt = earliest
	return result, nil
}

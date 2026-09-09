package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/mq"
)

const (
	// domainOutboxPending 表示领域事件已经提交、等待投递 Kafka。
	domainOutboxPending = "pending"
	// domainOutboxRunning 表示当前 worker 已通过租约抢占事件。
	domainOutboxRunning = "running"
	// domainOutboxSuccess 表示 Kafka 已确认接收事件。
	domainOutboxSuccess = "success"
	// domainOutboxFailed 表示超过自动重试上限，需要人工处理。
	domainOutboxFailed = "failed"
	// domainOutboxLeaseDuration 限制单次投递的持有时间，避免 worker 异常退出后任务永久卡住。
	domainOutboxLeaseDuration = 2 * time.Minute
	// domainOutboxDefaultMaxRetry 是未在任务记录中配置最大次数时的安全默认值。
	domainOutboxDefaultMaxRetry = 5
)

// AdminDomainOutbox 对齐 17_admin_domain_closure.sql 中的领域可靠事件表。
// job 只负责抢占、投递和更新投递状态，不修改处罚、退款、活动或发券等领域业务状态。
type AdminDomainOutbox struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	EventNo        string     `gorm:"column:event_no"`
	EventType      string     `gorm:"column:event_type"`
	AggregateType  string     `gorm:"column:aggregate_type"`
	AggregateID    int64      `gorm:"column:aggregate_id"`
	RequestID      string     `gorm:"column:request_id"`
	Payload        string     `gorm:"column:payload"`
	Status         string     `gorm:"column:status"`
	RetryCount     int        `gorm:"column:retry_count"`
	MaxRetry       int        `gorm:"column:max_retry"`
	NextRetryAt    time.Time  `gorm:"column:next_retry_at"`
	LeaseOwner     string     `gorm:"column:lease_owner"`
	LeaseExpiresAt *time.Time `gorm:"column:lease_expires_at"`
	FailureReason  string     `gorm:"column:failure_reason"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

// TableName 返回领域可靠事件的真实表名，禁止 GORM 自动迁移。
func (AdminDomainOutbox) TableName() string {
	return "admin_domain_outbox"
}

// domainOutboxEnvelope 是发送到 Kafka 的稳定事件信封。
// Payload 保持原始 JSON，避免 job 重新解释领域字段导致事件语义漂移。
type domainOutboxEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	RequestID     string          `json:"request_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

// RetryAdminDomainOutbox 扫描到期事件、基于租约抢占并投递到 Kafka。
// Kafka 未配置或发送失败时任务保留为 pending；超过最大次数才转为 failed。
func (t *Task) RetryAdminDomainOutbox(max int) error {
	if t == nil || t.svcCtx == nil || t.svcCtx.Db == nil {
		return errors.New("admin domain outbox task requires mysql")
	}
	if max <= 0 {
		max = 50
	}
	if len(t.svcCtx.Config.Kafka.Brokers) == 0 {
		return errors.New("admin domain outbox kafka is not configured")
	}
	now := time.Now()
	var items []AdminDomainOutbox
	if err := t.svcCtx.Db.WithContext(context.Background()).
		Where(
			"(status = ? AND next_retry_at <= ?) OR (status = ? AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)",
			domainOutboxPending, now, domainOutboxRunning, now,
		).
		Order("id ASC").
		Limit(max).
		Find(&items).Error; err != nil {
		return fmt.Errorf("查询管理后台领域outbox失败: %w", err)
	}

	for i := range items {
		if err := t.retryOneAdminDomainOutbox(context.Background(), &items[i]); err != nil {
			return err
		}
	}
	return nil
}

// retryOneAdminDomainOutbox 以条件更新抢占单条事件，防止多个 job 实例重复发送。
func (t *Task) retryOneAdminDomainOutbox(ctx context.Context, item *AdminDomainOutbox) error {
	if item == nil || item.ID <= 0 {
		return nil
	}
	now := time.Now()
	leaseOwner := fmt.Sprintf("job-%d", now.UnixNano())
	leaseExpiresAt := now.Add(domainOutboxLeaseDuration)
	claim := t.svcCtx.Db.WithContext(ctx).
		Model(&AdminDomainOutbox{}).
		Where(
			"id = ? AND ((status = ? AND next_retry_at <= ?) OR (status = ? AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?))",
			item.ID, domainOutboxPending, now, domainOutboxRunning, now,
		).
		Updates(map[string]any{
			"status":           domainOutboxRunning,
			"lease_owner":      leaseOwner,
			"lease_expires_at": leaseExpiresAt,
			"updated_at":       now,
		})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected == 0 {
		return nil
	}

	if err := t.publishAdminDomainOutbox(ctx, item); err != nil {
		return t.markAdminDomainOutboxFailed(ctx, item, leaseOwner, err)
	}
	return t.svcCtx.Db.WithContext(ctx).
		Model(&AdminDomainOutbox{}).
		Where("id = ? AND status = ? AND lease_owner = ?", item.ID, domainOutboxRunning, leaseOwner).
		Updates(map[string]any{
			"status":           domainOutboxSuccess,
			"lease_owner":      "",
			"lease_expires_at": nil,
			"failure_reason":   "",
			"updated_at":       time.Now(),
		}).Error
}

// publishAdminDomainOutbox 将单条事件转换为统一信封并投递到固定主题。
func (t *Task) publishAdminDomainOutbox(_ context.Context, item *AdminDomainOutbox) error {
	payload := json.RawMessage(`{}`)
	if strings.TrimSpace(item.Payload) != "" {
		payload = json.RawMessage(item.Payload)
		if !json.Valid(payload) {
			return errors.New("领域outbox载荷不是合法JSON")
		}
	}
	raw, err := json.Marshal(domainOutboxEnvelope{
		EventID:       item.EventNo,
		EventType:     item.EventType,
		AggregateType: item.AggregateType,
		AggregateID:   item.AggregateID,
		RequestID:     item.RequestID,
		OccurredAt:    item.CreatedAt,
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("序列化领域outbox事件失败: %w", err)
	}
	if t.svcCtx.EventProducer == nil {
		return errors.New("领域outbox Kafka生产者未配置")
	}
	if _, isNoopProducer := t.svcCtx.EventProducer.(*mq.NoopProducer); isNoopProducer {
		return errors.New("领域outbox Kafka生产者未就绪")
	}
	return t.svcCtx.EventProducer.Send(constants.TopicAdminDomain, item.EventNo, raw)
}

// markAdminDomainOutboxFailed 记录失败原因并按指数退避重新排期。
// 达到最大次数后置为 failed，保留失败原因和事件载荷供后台人工处理。
func (t *Task) markAdminDomainOutboxFailed(ctx context.Context, item *AdminDomainOutbox, leaseOwner string, cause error) error {
	retryCount := item.RetryCount + 1
	maxRetry := item.MaxRetry
	if maxRetry <= 0 {
		maxRetry = domainOutboxDefaultMaxRetry
	}
	status := domainOutboxPending
	nextRetryAt := time.Now().Add(domainOutboxRetryDelay(retryCount))
	if retryCount >= maxRetry {
		status = domainOutboxFailed
		nextRetryAt = time.Time{}
	}
	return t.svcCtx.Db.WithContext(ctx).
		Model(&AdminDomainOutbox{}).
		Where("id = ? AND status = ? AND lease_owner = ?", item.ID, domainOutboxRunning, leaseOwner).
		Updates(map[string]any{
			"status":           status,
			"retry_count":      retryCount,
			"next_retry_at":    nullableRetryTime(nextRetryAt),
			"lease_owner":      "",
			"lease_expires_at": nil,
			"failure_reason":   truncateDomainOutboxFailure(cause),
			"updated_at":       time.Now(),
		}).Error
}

// domainOutboxRetryDelay 返回 5 秒起步、上限 5 分钟的指数退避时长。
func domainOutboxRetryDelay(retryCount int) time.Duration {
	if retryCount <= 0 {
		return 5 * time.Second
	}
	seconds := 5 * math.Pow(2, float64(retryCount-1))
	if seconds > 300 {
		seconds = 300
	}
	return time.Duration(seconds) * time.Second
}

// nullableRetryTime 将终态任务的零时间写为 SQL NULL，避免被后续 pending 查询再次捞取。
func nullableRetryTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

// truncateDomainOutboxFailure 统一限制失败原因长度，防止失败状态回写再次因字段溢出失败。
func truncateDomainOutboxFailure(cause error) string {
	if cause == nil {
		return ""
	}
	message := cause.Error()
	if len(message) > 512 {
		return message[:512]
	}
	return message
}

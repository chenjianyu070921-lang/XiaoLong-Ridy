package adminservicelogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
)

const (
	// domainOutboxStatusPending 表示事件已随本地事务提交、等待 job 抢占投递。
	domainOutboxStatusPending = "pending"
	// domainOutboxDefaultMaxRetry 统一限制后台领域事件的最大自动投递次数。
	domainOutboxDefaultMaxRetry = 5
)

// AdminDomainEvent 表示写入 admin_domain_outbox 的最小可靠事件。
// Payload 只能保存下游执行所需的业务字段，禁止携带口令、会话令牌和完整敏感身份信息。
type AdminDomainEvent struct {
	EventType     string
	AggregateType string
	AggregateID   int64
	RequestID     string
	Payload       any
}

// recordDomainOutboxTx 在调用方已开启的本地事务内写入领域事件。
// 入参 tx 必须与业务状态和 admin_operation_log 使用同一事务，返回 eventNo 供审计和下游幂等追踪。
func recordDomainOutboxTx(ctx context.Context, tx *sql.Tx, event AdminDomainEvent) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("领域outbox事务不能为空")
	}
	if err := validateDomainEvent(event); err != nil {
		return "", err
	}
	payload, err := marshalSafeDomainPayload(event.Payload)
	if err != nil {
		return "", err
	}
	eventNo := newAdminTaskNo("ADO")
	_, err = tx.ExecContext(ctx, `
		INSERT INTO admin_domain_outbox (
			event_no, event_type, aggregate_type, aggregate_id, request_id, payload,
			status, retry_count, max_retry, next_retry_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, NOW())
	`, eventNo, event.EventType, event.AggregateType, event.AggregateID, event.RequestID,
		payload, domainOutboxStatusPending, domainOutboxDefaultMaxRetry)
	if err != nil {
		return "", err
	}
	return eventNo, nil
}

// validateDomainEvent 在访问数据库前校验事件的聚合、幂等键和事件类型。
func validateDomainEvent(event AdminDomainEvent) error {
	if strings.TrimSpace(event.EventType) == "" {
		return fmt.Errorf("领域事件类型不能为空")
	}
	if strings.TrimSpace(event.AggregateType) == "" || event.AggregateID <= 0 {
		return fmt.Errorf("领域事件聚合对象不合法")
	}
	if strings.TrimSpace(event.RequestID) == "" {
		return fmt.Errorf("领域事件request_id不能为空")
	}
	return nil
}

// marshalSafeDomainPayload 序列化领域事件载荷并阻止常见敏感字段进入消息表。
// 该校验不是脱敏替代方案；调用方仍必须只构造满足领域最小化原则的 payload。
func marshalSafeDomainPayload(payload any) ([]byte, error) {
	if payload == nil {
		return []byte(`{}`), nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化领域事件载荷失败: %w", err)
	}
	var fields any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("解析领域事件载荷失败: %w", err)
	}
	if containsSensitiveDomainField(fields) {
		return nil, fmt.Errorf("领域事件载荷包含敏感字段")
	}
	return raw, nil
}

// containsSensitiveDomainField 递归检查 JSON 字段名，阻止明显的认证与身份敏感数据进入 outbox。
func containsSensitiveDomainField(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
			switch normalized {
			case "password", "passwordhash", "token", "authorization", "idcardno", "bankcardno":
				return true
			}
			if containsSensitiveDomainField(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsSensitiveDomainField(nested) {
				return true
			}
		}
	}
	return false
}

// recordDomainOutboxAfterCommitted 是跨服务操作已成功后的兜底记录入口。
// 第一阶段写操作应优先使用 recordDomainOutboxTx；该函数只为无法纳入同一事务的旧链路保留。
func recordDomainOutboxAfterCommitted(ctx context.Context, svcCtx *svc.ServiceContext, event AdminDomainEvent) (string, error) {
	if svcCtx == nil || svcCtx.MySQL == nil {
		return "", fmt.Errorf("领域outbox数据库未配置")
	}
	tx, err := svcCtx.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	eventNo, err := recordDomainOutboxTx(ctx, tx, event)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return eventNo, nil
}

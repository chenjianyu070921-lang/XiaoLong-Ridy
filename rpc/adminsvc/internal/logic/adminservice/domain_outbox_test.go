package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRecordDomainOutboxTx 验证领域事件以调用方事务写入，并携带统一的初始投递状态。
func TestRecordDomainOutboxTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建sqlmock失败: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("开启事务失败: %v", err)
	}
	mock.ExpectExec(`INSERT INTO admin_domain_outbox`).
		WithArgs(sqlmock.AnyArg(), "admin.driver.punishment.requested", "driver_punishment", int64(101), "req-101",
			sqlmock.AnyArg(), domainOutboxStatusPending, domainOutboxDefaultMaxRetry).
		WillReturnResult(sqlmock.NewResult(1, 1))

	eventNo, err := recordDomainOutboxTx(context.Background(), tx, AdminDomainEvent{
		EventType: "admin.driver.punishment.requested", AggregateType: "driver_punishment",
		AggregateID: 101, RequestID: "req-101", Payload: map[string]any{"driver_id": 1001},
	})
	if err != nil {
		t.Fatalf("写入领域outbox失败: %v", err)
	}
	if eventNo == "" {
		t.Fatal("领域事件编号不能为空")
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("回滚事务失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock预期未满足: %v", err)
	}
}

// TestMarshalSafeDomainPayloadRejectsSensitiveFields 验证密码和令牌等敏感字段不能写入可靠事件表。
func TestMarshalSafeDomainPayloadRejectsSensitiveFields(t *testing.T) {
	for _, payload := range []map[string]any{
		{"token": "secret"},
		{"operator": map[string]any{"password_hash": "secret"}},
		{"items": []any{map[string]any{"bank_card_no": "6222"}}},
	} {
		if _, err := marshalSafeDomainPayload(payload); err == nil {
			t.Fatalf("敏感字段载荷应被拒绝: %#v", payload)
		}
	}
}

// TestValidateDomainEvent 验证事件的聚合对象和幂等键均为必填条件。
func TestValidateDomainEvent(t *testing.T) {
	if err := validateDomainEvent(AdminDomainEvent{EventType: "x", AggregateType: "a", AggregateID: 1, RequestID: "r"}); err != nil {
		t.Fatalf("合法事件校验失败: %v", err)
	}
	if err := validateDomainEvent(AdminDomainEvent{EventType: "x", AggregateType: "a", AggregateID: 1}); err == nil {
		t.Fatal("缺少request_id的事件应被拒绝")
	}
}

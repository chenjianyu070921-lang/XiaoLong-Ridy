package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/proto"
	pushproto "XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	adminOutboxStatusPending = "pending"
	adminOutboxStatusRunning = "running"
	adminOutboxStatusSuccess = "success"
	adminOutboxStatusFailed  = "failed"
	defaultAdminOutboxMax    = 5
)

// AdminAuditOutbox 表示管理后台审计补偿任务。
// 该模型严格对齐 scripts/sql/migrate/09_admin_export_audit_task.sql，不通过 AutoMigrate 修改数据库结构。
type AdminAuditOutbox struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	EventNo       string    `gorm:"column:event_no"`
	Module        string    `gorm:"column:module"`
	Action        string    `gorm:"column:action"`
	TargetType    string    `gorm:"column:target_type"`
	TargetID      int64     `gorm:"column:target_id"`
	AdminID       int64     `gorm:"column:admin_id"`
	Detail        string    `gorm:"column:detail"`
	IP            string    `gorm:"column:ip"`
	Status        string    `gorm:"column:status"`
	RetryCount    int       `gorm:"column:retry_count"`
	FailureReason string    `gorm:"column:failure_reason"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName 返回管理后台 outbox 真实表名。
func (AdminAuditOutbox) TableName() string {
	return "admin_audit_outbox"
}

// RetryAdminAuditOutbox 扫描并补偿管理后台 outbox。
// max 限制单轮处理条数，防止积压时长时间占用 job 调度线程。
func (t *Task) RetryAdminAuditOutbox(max int) error {
	if t == nil || t.svcCtx == nil || t.svcCtx.Db == nil {
		return errors.New("admin audit outbox task requires mysql")
	}
	if max <= 0 {
		max = 50
	}
	maxRetry := t.svcCtx.Config.AdminOutboxMaxRetry
	if maxRetry <= 0 {
		maxRetry = defaultAdminOutboxMax
	}
	ctx := context.Background()
	var items []AdminAuditOutbox
	staleRunningBefore := time.Now().Add(-5 * time.Minute)
	if err := t.svcCtx.Db.WithContext(ctx).
		Where("retry_count < ? AND (status = ? OR (status = ? AND updated_at < ?))", maxRetry, adminOutboxStatusPending, adminOutboxStatusRunning, staleRunningBefore).
		Order("id ASC").
		Limit(max).
		Find(&items).Error; err != nil {
		return fmt.Errorf("查询管理后台 outbox 失败: %w", err)
	}
	successCount := 0
	for i := range items {
		if err := t.retryOneAdminAuditOutbox(ctx, &items[i], maxRetry); err != nil {
			logx.Errorf("管理后台 outbox 补偿失败 event_no=%s action=%s target=%s/%d: %v", items[i].EventNo, items[i].Action, items[i].TargetType, items[i].TargetID, err)
			continue
		}
		successCount++
	}
	logx.Infof("管理后台 outbox 扫描完成: 拉取 %d 条, 成功 %d 条", len(items), successCount)
	return nil
}

// retryOneAdminAuditOutbox 抢占并执行单条 outbox。
// 通过 id + status 条件更新为 running，避免多个 job 实例重复处理同一条补偿任务。
func (t *Task) retryOneAdminAuditOutbox(ctx context.Context, item *AdminAuditOutbox, maxRetry int) error {
	if item == nil || item.ID <= 0 {
		return nil
	}
	claim := t.svcCtx.Db.WithContext(ctx).
		Model(&AdminAuditOutbox{}).
		Where("id = ? AND status IN ?", item.ID, []string{adminOutboxStatusPending, adminOutboxStatusRunning}).
		Updates(map[string]any{"status": adminOutboxStatusRunning, "updated_at": time.Now()})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected == 0 {
		return nil
	}
	if err := t.executeAdminAuditOutbox(ctx, item); err != nil {
		return t.markAdminOutboxFailed(ctx, item, maxRetry, err)
	}
	return t.svcCtx.Db.WithContext(ctx).
		Model(&AdminAuditOutbox{}).
		Where("id = ?", item.ID).
		Updates(map[string]any{"status": adminOutboxStatusSuccess, "failure_reason": "", "updated_at": time.Now()}).Error
}

// executeAdminAuditOutbox 根据 module/action 执行具体补偿动作。
// 普通审计事件重写 admin_operation_log；司机冻结和通知事件分别重放 driversvc 与 pushsvc。
func (t *Task) executeAdminAuditOutbox(ctx context.Context, item *AdminAuditOutbox) error {
	action := strings.TrimSpace(item.Action)
	switch {
	case action == "freeze_driver":
		return t.replayDriverFreeze(ctx, item)
	case strings.HasSuffix(action, "_notice"):
		return t.replayDriverNotice(ctx, item)
	case strings.HasSuffix(action, "_push"):
		return t.replayDriverPush(ctx, item)
	case strings.HasSuffix(action, "_notify"):
		if err := t.replayDriverNotice(ctx, item); err != nil {
			return err
		}
		return t.replayDriverPush(ctx, item)
	default:
		return t.replayOperationLog(ctx, item)
	}
}

// replayOperationLog 将审计补偿事件重新写入 admin_operation_log。
func (t *Task) replayOperationLog(ctx context.Context, item *AdminAuditOutbox) error {
	return t.svcCtx.Db.WithContext(ctx).Exec(`
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, item.AdminID, item.Module, item.Action, item.TargetType, item.TargetID, item.Detail, item.IP).Error
}

// replayDriverFreeze 重放司机冻结补偿。
// 该动作只通过 driversvc 修改司机域状态，job 不直接更新司机表。
func (t *Task) replayDriverFreeze(ctx context.Context, item *AdminAuditOutbox) error {
	if t.svcCtx.DriverClient == nil {
		return errors.New("driversvc client is not configured")
	}
	_, err := t.svcCtx.DriverClient.FreezeDriver(ctx, &proto.FreezeDriverRequest{
		DriverId:   item.TargetID,
		Reason:     outboxDetailOrDefault(item.Detail, "风控黑名单补偿冻结司机"),
		OperatorId: item.AdminID,
		Ip:         item.IP,
	})
	return err
}

// replayDriverNotice 重放司机站内信通知补偿。
func (t *Task) replayDriverNotice(ctx context.Context, item *AdminAuditOutbox) error {
	if t.svcCtx.PushClient == nil {
		return errors.New("pushsvc client is not configured")
	}
	_, err := t.svcCtx.PushClient.SendNotice(ctx, &pushproto.SendNoticeReq{
		UserId:  item.TargetID,
		Title:   "后台操作通知",
		Content: outboxDetailOrDefault(item.Detail, "后台操作状态已更新"),
		BizType: 2,
	})
	return err
}

// replayDriverPush 重放司机 App 推送补偿。
func (t *Task) replayDriverPush(ctx context.Context, item *AdminAuditOutbox) error {
	if t.svcCtx.PushClient == nil {
		return errors.New("pushsvc client is not configured")
	}
	_, err := t.svcCtx.PushClient.SendPush(ctx, &pushproto.SendPushReq{
		UserId:     item.TargetID,
		Title:      "后台操作通知",
		Body:       outboxDetailOrDefault(item.Detail, "后台操作状态已更新"),
		Extras:     fmt.Sprintf(`{"target_type":"%s","target_id":%d,"action":"%s","event_no":"%s"}`, item.TargetType, item.TargetID, item.Action, item.EventNo),
		DeviceType: "driver",
	})
	return err
}

// markAdminOutboxFailed 记录补偿失败结果。
// 未超过最大次数时回到 pending 等待下一轮；超过后置为 failed，保留 failure_reason 供人工处理。
func (t *Task) markAdminOutboxFailed(ctx context.Context, item *AdminAuditOutbox, maxRetry int, cause error) error {
	retryCount := item.RetryCount + 1
	status := adminOutboxStatusPending
	if retryCount >= maxRetry {
		status = adminOutboxStatusFailed
	}
	return t.svcCtx.Db.WithContext(ctx).
		Model(&AdminAuditOutbox{}).
		Where("id = ?", item.ID).
		Updates(map[string]any{
			"status":         status,
			"retry_count":    retryCount,
			"failure_reason": truncateOutboxFailure(cause),
			"updated_at":     time.Now(),
		}).Error
}

// outboxDetailOrDefault 返回补偿事件详情，空值时使用默认文案。
func outboxDetailOrDefault(detail, fallback string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return fallback
	}
	return detail
}

// truncateOutboxFailure 将失败原因限制到表字段长度内，避免补偿写回再次失败。
func truncateOutboxFailure(cause error) string {
	if cause == nil {
		return ""
	}
	msg := cause.Error()
	if len(msg) > 512 {
		return msg[:512]
	}
	return msg
}

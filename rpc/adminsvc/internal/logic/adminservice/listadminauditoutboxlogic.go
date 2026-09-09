package adminservicelogic

import (
	"context"
	"database/sql"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAdminAuditOutboxLogic 负责查询管理后台通知与审计补偿任务。
// 该查询只读取 admin_audit_outbox，不直接执行补偿；任务仍由 job 统一抢占和重试。
type ListAdminAuditOutboxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAdminAuditOutboxLogic 创建补偿任务查询逻辑对象。
func NewListAdminAuditOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminAuditOutboxLogic {
	return &ListAdminAuditOutboxLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListAdminAuditOutbox 按状态、模块、动作和目标查询补偿任务，返回失败原因与重试次数供运营追踪。
func (l *ListAdminAuditOutboxLogic) ListAdminAuditOutbox(in *adminsvc.AdminAuditOutboxListRequest) (*adminsvc.AdminAuditOutboxListResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2); err != nil {
		return nil, err
	}
	if l.svcCtx == nil || l.svcCtx.MySQL == nil {
		return nil, status.Error(codes.FailedPrecondition, "mysql client not ready")
	}

	where := " WHERE 1=1"
	args := make([]any, 0, 4)
	if value := strings.TrimSpace(in.GetStatus()); value != "" {
		where += " AND status = ?"
		args = append(args, value)
	}
	if value := strings.TrimSpace(in.GetModule()); value != "" {
		where += " AND module = ?"
		args = append(args, value)
	}
	if value := strings.TrimSpace(in.GetAction()); value != "" {
		where += " AND action = ?"
		args = append(args, value)
	}
	if in.GetTargetId() > 0 {
		where += " AND target_id = ?"
		args = append(args, in.GetTargetId())
	}

	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, "SELECT COUNT(1) FROM admin_audit_outbox"+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	page := normalizePage(in.GetPage())
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(append([]any{}, args...), limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, event_no, module, action, target_type, target_id, admin_id, detail,
		       status, retry_count, failure_reason, created_at, updated_at
		FROM admin_audit_outbox`+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*adminsvc.AdminAuditOutbox, 0)
	for rows.Next() {
		item := &adminsvc.AdminAuditOutbox{}
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.Id, &item.EventNo, &item.Module, &item.Action, &item.TargetType,
			&item.TargetId, &item.AdminId, &item.Detail, &item.Status, &item.RetryCount,
			&item.FailureReason, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatNullTime(createdAt)
		item.UpdatedAt = formatNullTime(updatedAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &adminsvc.AdminAuditOutboxListResponse{
		List: items, Total: total, Page: page, PageSize: limit,
	}, nil
}

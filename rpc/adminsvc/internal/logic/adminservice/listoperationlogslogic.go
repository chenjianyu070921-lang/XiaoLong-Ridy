package adminservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListOperationLogsLogic 处理后台操作日志列表查询 RPC。
type ListOperationLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListOperationLogsLogic 创建操作日志列表查询逻辑对象。
func NewListOperationLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOperationLogsLogic {
	return &ListOperationLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOperationLogs 按分页和筛选条件查询后台操作日志。
func (l *ListOperationLogsLogic) ListOperationLogs(in *adminsvc.OperationLogListRequest) (*adminsvc.OperationLogListResponse, error) {
	where, args := buildOperationLogWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM admin_operation_log `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, admin_id, module, action, target_type, target_id, detail, ip, created_at
		FROM admin_operation_log `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.OperationLog, 0)
	for rows.Next() {
		var item adminsvc.OperationLog
		var createdAt sql.NullTime
		if err := rows.Scan(&item.Id, &item.AdminId, &item.Module, &item.Action, &item.TargetType, &item.TargetId, &item.Detail, &item.Ip, &createdAt); err != nil {
			return nil, fmt.Errorf("scan operation log row: %w", err)
		}
		item.CreatedAt = formatNullTime(createdAt)
		list = append(list, &item)
	}
	return &adminsvc.OperationLogListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// buildOperationLogWhere 组装操作日志查询条件。
func buildOperationLogWhere(in *adminsvc.OperationLogListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetAdminId() > 0 {
		parts = append(parts, "admin_id = ?")
		args = append(args, in.GetAdminId())
	}
	if in.GetModule() != "" {
		parts = append(parts, "module = ?")
		args = append(args, in.GetModule())
	}
	if in.GetAction() != "" {
		parts = append(parts, "action = ?")
		args = append(args, in.GetAction())
	}
	if in.GetTargetType() != "" {
		parts = append(parts, "target_type = ?")
		args = append(args, in.GetTargetType())
	}
	if in.GetTargetId() > 0 {
		parts = append(parts, "target_id = ?")
		args = append(args, in.GetTargetId())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

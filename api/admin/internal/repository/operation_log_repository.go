package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// ErrOperationLogNotFound 表示没有找到操作日志。
var ErrOperationLogNotFound = errors.New("operation log not found")

// OperationLogRepository 封装后台操作日志表 admin_operation_log 的读写。
type OperationLogRepository struct {
	db *sql.DB
}

// NewOperationLogRepository 创建操作日志仓储。
func NewOperationLogRepository(db *sql.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create 写入一条后台操作日志。
func (r *OperationLogRepository) Create(ctx context.Context, input CreateOperationLogInput) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.AdminID, input.Module, input.Action, input.TargetType, input.TargetID, input.Detail, input.IP)
	return err
}

// List 查询后台操作日志列表。
func (r *OperationLogRepository) List(ctx context.Context, req types.OperationLogListRequest) ([]model.OperationLog, int64, error) {
	where, args := buildOperationLogWhere(req)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM admin_operation_log `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, admin_id, module, action, target_type, target_id, detail, ip, created_at
		FROM admin_operation_log `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.OperationLog, 0)
	for rows.Next() {
		item, err := scanOperationLogRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}

// buildOperationLogWhere 组装操作日志查询条件。
func buildOperationLogWhere(req types.OperationLogListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if req.AdminID > 0 {
		parts = append(parts, "admin_id = ?")
		args = append(args, req.AdminID)
	}
	if req.Module != "" {
		parts = append(parts, "module = ?")
		args = append(args, req.Module)
	}
	if req.Action != "" {
		parts = append(parts, "action = ?")
		args = append(args, req.Action)
	}
	if req.TargetType != "" {
		parts = append(parts, "target_type = ?")
		args = append(args, req.TargetType)
	}
	if req.TargetID > 0 {
		parts = append(parts, "target_id = ?")
		args = append(args, req.TargetID)
	}
	if req.StartTime != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, req.EndTime)
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// scanOperationLogRows 扫描操作日志列表结果。
func scanOperationLogRows(rows *sql.Rows) (*model.OperationLog, error) {
	var item model.OperationLog
	if err := rows.Scan(&item.ID, &item.AdminID, &item.Module, &item.Action, &item.TargetType, &item.TargetID, &item.Detail, &item.IP, &item.CreatedAt); err != nil {
		return nil, fmt.Errorf("scan operation log row: %w", err)
	}
	return &item, nil
}

// CreateOperationLogInput 表示操作日志写入参数。
type CreateOperationLogInput struct {
	AdminID    int64
	Module     string
	Action     string
	TargetType string
	TargetID   int64
	Detail     string
	IP         string
}

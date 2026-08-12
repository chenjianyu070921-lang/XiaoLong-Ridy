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

// ListUsersLogic 处理用户列表查询 RPC。
type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListUsersLogic 创建用户列表查询逻辑对象。
func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListUsers 按分页和筛选条件查询乘客用户列表。
func (l *ListUsersLogic) ListUsers(in *adminsvc.UserListRequest) (*adminsvc.UserListResponse, error) {
	where, args := buildUserWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM user `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, phone, nickname, avatar_url, gender, real_name, id_card_no,
		       register_source, status, created_at, updated_at
		FROM user `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.User, 0)
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.UserListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// buildUserWhere 组装用户列表筛选条件。
func buildUserWhere(in *adminsvc.UserListRequest) (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "(phone LIKE ? OR nickname LIKE ? OR real_name LIKE ?)")
		kw := "%" + in.GetKeyword() + "%"
		args = append(args, kw, kw, kw)
	}
	if in.GetStatus() > 0 {
		parts = append(parts, "status = ?")
		args = append(args, in.GetStatus())
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

// scanUser 将用户查询结果转换为 protobuf 用户对象。
func scanUser(rows *sql.Rows) (*adminsvc.User, error) {
	var item adminsvc.User
	var createdAt, updatedAt sql.NullTime
	if err := rows.Scan(
		&item.Id, &item.Phone, &item.Nickname, &item.AvatarUrl, &item.Gender,
		&item.RealName, &item.IdCardNo, &item.RegisterSource, &item.Status,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan user row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

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

// ErrUserNotFound 表示没有找到指定用户。
var ErrUserNotFound = errors.New("user not found")

// UserRepository 封装 user 表的后台查询能力。
// P0 阶段只提供列表和详情查询，不在后台直接修改用户业务数据。
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储。
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// List 按分页和筛选条件查询用户列表。
func (r *UserRepository) List(ctx context.Context, req types.UserListRequest) ([]model.User, int64, error) {
	where, args := buildUserWhere(req)
	countSQL := `SELECT COUNT(1) FROM user ` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, phone, password_hash, nickname, avatar_url, gender, real_name, id_card_no,
		       register_source, status, created_at, updated_at, deleted_at
		FROM user `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.User, 0)
	for rows.Next() {
		item, err := scanUserRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}

// GetByID 根据用户 ID 查询用户详情。
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, phone, password_hash, nickname, avatar_url, gender, real_name, id_card_no,
		       register_source, status, created_at, updated_at, deleted_at
		FROM user
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	var user model.User
	var deletedAt sql.NullTime
	err := row.Scan(
		&user.ID, &user.Phone, &user.PasswordHash, &user.Nickname, &user.AvatarURL,
		&user.Gender, &user.RealName, &user.IDCardNo, &user.RegisterSource,
		&user.Status, &user.CreatedAt, &user.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}
	return &user, nil
}

// buildUserWhere 组装用户列表筛选条件。
func buildUserWhere(req types.UserListRequest) (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if req.Keyword != "" {
		parts = append(parts, "(phone LIKE ? OR nickname LIKE ? OR real_name LIKE ?)")
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if req.Status > 0 {
		parts = append(parts, "status = ?")
		args = append(args, req.Status)
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

// scanUserRows 从列表查询结果中扫描用户模型。
func scanUserRows(rows *sql.Rows) (*model.User, error) {
	var user model.User
	var deletedAt sql.NullTime
	err := rows.Scan(
		&user.ID, &user.Phone, &user.PasswordHash, &user.Nickname, &user.AvatarURL,
		&user.Gender, &user.RealName, &user.IDCardNo, &user.RegisterSource,
		&user.Status, &user.CreatedAt, &user.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan user row: %w", err)
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}
	return &user, nil
}

// normalizePageSize 统一限制分页大小，避免后台一次拉取过多数据。
func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

// normalizePage 统一页码默认值。
func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizeOffset 根据页码和每页条数计算 SQL offset。
func normalizeOffset(page, pageSize int) int {
	return (normalizePage(page) - 1) * normalizePageSize(pageSize)
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/api/admin/internal/model"
)

var (
	// ErrAdminNotFound 表示没有找到指定的管理员账号。
	ErrAdminNotFound = errors.New("admin not found")
)

// AdminRepository 封装 admin_user 表的管理员账号读写操作。
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository 创建管理员账号数据仓储。
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// Count 统计未被软删除的管理员账号数量，用于判断是否允许首次注册。
func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM admin_user WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

// GetByUsername 按用户名查询有效的管理员账号。
func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, real_name, role, status, last_login_at, created_at, updated_at, deleted_at
		FROM admin_user
		WHERE username = ? AND deleted_at IS NULL
	`, username)
	return scanAdminUser(row)
}

// GetByID 按主键查询有效的管理员账号。
func (r *AdminRepository) GetByID(ctx context.Context, id int64) (*model.AdminUser, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, real_name, role, status, last_login_at, created_at, updated_at, deleted_at
		FROM admin_user
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	return scanAdminUser(row)
}

// Create 新建管理员账号，并返回数据库生成的主键。
func (r *AdminRepository) Create(ctx context.Context, input CreateAdminInput) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_user (username, password_hash, real_name, role, status, last_login_at, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Username, input.PasswordHash, input.RealName, input.Role, input.Status, nil, now, now, nil)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateLastLoginAt 更新管理员最近一次登录时间。
func (r *AdminRepository) UpdateLastLoginAt(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admin_user SET last_login_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`, at, at, id)
	return err
}

// scanAdminUser 将查询结果行转换为管理员模型，并处理可空时间字段。
func scanAdminUser(row *sql.Row) (*model.AdminUser, error) {
	var admin model.AdminUser
	var lastLoginAt, deletedAt sql.NullTime
	err := row.Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.RealName,
		&admin.Role,
		&admin.Status,
		&lastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminNotFound
		}
		return nil, fmt.Errorf("scan admin user: %w", err)
	}
	if lastLoginAt.Valid {
		admin.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		admin.DeletedAt = &deletedAt.Time
	}
	return &admin, nil
}

// CreateAdminInput 定义创建管理员账号所需的持久化字段。
type CreateAdminInput struct {
	Username     string
	PasswordHash string
	RealName     string
	Role         int32
	Status       int32
}

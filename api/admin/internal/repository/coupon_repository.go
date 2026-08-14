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

// ErrCouponNotFound 表示没有找到指定优惠券模板。
// logic 层会把该错误转换为统一的 404 业务响应。
var ErrCouponNotFound = errors.New("coupon not found")

// CouponRepository 封装 coupon 表的数据访问逻辑。
// 管理后台当前只负责模板配置，不在这里直接发放用户券，避免影响用户侧业务数据。
type CouponRepository struct {
	db *sql.DB
}

// NewCouponRepository 创建优惠券仓储实例。
// 入参 db 为管理后台启动时初始化的 MySQL 连接池。
func NewCouponRepository(db *sql.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// List 按分页和筛选条件查询优惠券模板列表。
// 返回值包含当前页模板数据、满足条件的总记录数和底层错误。
func (r *CouponRepository) List(ctx context.Context, req types.CouponListRequest) ([]model.Coupon, int64, error) {
	where, args := buildCouponWhere(req)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM coupon `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, type, CAST(face_value AS CHAR), CAST(discount AS CHAR),
		       CAST(threshold_amount AS CHAR), total_count, received_count, per_user_limit,
		       valid_start_at, valid_end_at, status, created_at, updated_at
		FROM coupon `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.Coupon, 0)
	for rows.Next() {
		item, err := scanCouponRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}

// Create 新增优惠券模板。
// 金额类字段由 MySQL DECIMAL 保存，调用方以字符串传入，避免 Go float 精度误差。
func (r *CouponRepository) Create(ctx context.Context, req types.CouponSaveRequest) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO coupon (
			name, type, face_value, discount, threshold_amount, total_count,
			received_count, per_user_limit, valid_start_at, valid_end_at, status
		) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
	`, req.Name, req.Type, req.FaceValue, req.Discount, req.ThresholdAmount, req.TotalCount,
		req.PerUserLimit, req.ValidStartAt, req.ValidEndAt, req.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update 编辑优惠券模板。
// 已领取数量不在后台编辑接口中更新，确保运营配置不会污染用户领券数据。
func (r *CouponRepository) Update(ctx context.Context, id int64, req types.CouponSaveRequest) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE coupon
		SET name = ?, type = ?, face_value = ?, discount = ?, threshold_amount = ?,
		    total_count = ?, per_user_limit = ?, valid_start_at = ?, valid_end_at = ?, status = ?
		WHERE id = ?
	`, req.Name, req.Type, req.FaceValue, req.Discount, req.ThresholdAmount,
		req.TotalCount, req.PerUserLimit, req.ValidStartAt, req.ValidEndAt, req.Status, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrCouponNotFound
	}
	return nil
}

// buildCouponWhere 组装优惠券模板列表的动态筛选条件。
// 所有外部输入均通过占位符传参，避免 SQL 注入风险。
func buildCouponWhere(req types.CouponListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if req.Keyword != "" {
		parts = append(parts, "name LIKE ?")
		args = append(args, "%"+req.Keyword+"%")
	}
	if req.Type > 0 {
		parts = append(parts, "type = ?")
		args = append(args, req.Type)
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

// scanCouponRows 将 SQL 查询结果转换为优惠券模板模型。
// 该函数集中处理字段扫描，便于后续新增字段时统一维护。
func scanCouponRows(rows *sql.Rows) (*model.Coupon, error) {
	var item model.Coupon
	err := rows.Scan(
		&item.ID, &item.Name, &item.Type, &item.FaceValue, &item.Discount,
		&item.ThresholdAmount, &item.TotalCount, &item.ReceivedCount, &item.PerUserLimit,
		&item.ValidStartAt, &item.ValidEndAt, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan coupon row: %w", err)
	}
	return &item, nil
}

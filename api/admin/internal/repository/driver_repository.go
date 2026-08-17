package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// ErrDriverCertificationNotFound 表示没有找到指定的司机认证审核记录。
var ErrDriverCertificationNotFound = errors.New("driver certification not found")

// DriverRepository 封装司机审核相关的数据访问。
// 当前项目尚未提供独立的 driversvc，因此 P0 阶段由管理后台直接查询相关业务表。
type DriverRepository struct {
	db *sql.DB
}

// NewDriverRepository 创建司机审核数据仓储。
func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

// ListCertifications 查询司机认证审核列表，并返回分页总数。
// 查询会关联司机和车辆信息，便于前端在同一列表展示审核摘要。
func (r *DriverRepository) ListCertifications(ctx context.Context, req types.DriverCertificationListRequest) ([]model.DriverCertification, int64, error) {
	where, args := buildCertificationWhere(req)
	countSQL := `
		SELECT COUNT(1)
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
	` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.driver_id, c.vehicle_id,
		       COALESCE(d.phone, ''), COALESCE(d.real_name, ''), COALESCE(d.status, 0),
		       COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       c.id_card_front_url, c.id_card_back_url, c.driver_license_url, c.vehicle_license_url,
		       c.audit_status, c.audit_remark, c.audited_by, c.audited_at, c.created_at, c.updated_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
	`+where+`
		ORDER BY c.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.DriverCertification, 0)
	for rows.Next() {
		item, err := scanCertificationRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}

// GetCertificationByID 根据主键查询单条司机认证审核详情。
func (r *DriverRepository) GetCertificationByID(ctx context.Context, id int64) (*model.DriverCertification, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.driver_id, c.vehicle_id,
		       COALESCE(d.phone, ''), COALESCE(d.real_name, ''), COALESCE(d.status, 0),
		       COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       c.id_card_front_url, c.id_card_back_url, c.driver_license_url, c.vehicle_license_url,
		       c.audit_status, c.audit_remark, c.audited_by, c.audited_at, c.created_at, c.updated_at
		FROM driver_certification c
		LEFT JOIN driver d ON d.id = c.driver_id
		LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id
		WHERE c.id = ?
	`, id)
	item, err := scanCertificationRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDriverCertificationNotFound
		}
		return nil, err
	}
	return item, nil
}

// AuditCertification 在一个事务中更新审核结果。
// 审核通过时同步激活司机和车辆；驳回时只更新审核记录，保留司机待审核状态。
func (r *DriverRepository) AuditCertification(ctx context.Context, id int64, auditStatus int32, remark string, adminID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Rollback 在 Commit 后只会返回 sql.ErrTxDone，因此可用于覆盖所有异常返回路径。
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	res, err := tx.ExecContext(ctx, `
		UPDATE driver_certification
		SET audit_status = ?, audit_remark = ?, audited_by = ?, audited_at = ?, updated_at = ?
		WHERE id = ?
	`, auditStatus, remark, adminID, now, now, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrDriverCertificationNotFound
	}

	if auditStatus == 2 {
		var driverID, vehicleID int64
		if err = tx.QueryRowContext(ctx, `SELECT driver_id, vehicle_id FROM driver_certification WHERE id = ?`, id).Scan(&driverID, &vehicleID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE driver SET status = 2, updated_at = ? WHERE id = ?`, now, driverID); err != nil {
			return err
		}
		if vehicleID > 0 {
			if _, err = tx.ExecContext(ctx, `UPDATE driver_vehicle SET status = 2, updated_at = ? WHERE id = ?`, now, vehicleID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// buildCertificationWhere 组装司机审核列表的动态筛选条件。
func buildCertificationWhere(req types.DriverCertificationListRequest) (string, []any) {
	parts := make([]string, 0)
	args := make([]any, 0)
	if req.Keyword != "" {
		parts = append(parts, "(d.phone LIKE ? OR d.real_name LIKE ? OR v.plate_no LIKE ?)")
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if req.AuditStatus > 0 {
		parts = append(parts, "c.audit_status = ?")
		args = append(args, req.AuditStatus)
	}
	if req.StartTime != "" {
		parts = append(parts, "c.created_at >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		parts = append(parts, "c.created_at <= ?")
		args = append(args, req.EndTime)
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

// scanCertificationRows 将列表查询的当前行转换为司机认证模型。
func scanCertificationRows(rows *sql.Rows) (*model.DriverCertification, error) {
	var item model.DriverCertification
	var auditedAt sql.NullTime
	err := rows.Scan(
		&item.ID, &item.DriverID, &item.VehicleID,
		&item.DriverPhone, &item.DriverName, &item.DriverStatus,
		&item.PlateNo, &item.VehicleStatus,
		&item.IDCardFrontURL, &item.IDCardBackURL, &item.DriverLicenseURL, &item.VehicleLicenseURL,
		&item.AuditStatus, &item.AuditRemark, &item.AuditedBy, &auditedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan certification row: %w", err)
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	return &item, nil
}

// scanCertificationRow 将单条查询结果转换为司机认证模型。
func scanCertificationRow(row *sql.Row) (*model.DriverCertification, error) {
	var item model.DriverCertification
	var auditedAt sql.NullTime
	err := row.Scan(
		&item.ID, &item.DriverID, &item.VehicleID,
		&item.DriverPhone, &item.DriverName, &item.DriverStatus,
		&item.PlateNo, &item.VehicleStatus,
		&item.IDCardFrontURL, &item.IDCardBackURL, &item.DriverLicenseURL, &item.VehicleLicenseURL,
		&item.AuditStatus, &item.AuditRemark, &item.AuditedBy, &auditedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan certification: %w", err)
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	return &item, nil
}

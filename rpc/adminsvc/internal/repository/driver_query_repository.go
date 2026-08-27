package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
)

var ErrDriverNotFound = errors.New("driver not found")

type DriverListFilter struct {
	Page     int32
	PageSize int32
	Keyword  string
	Status   int32
}

type DriverListResult struct {
	List     []*adminsvc.Driver
	Total    int64
	Page     int32
	PageSize int32
}

type DriverQueryRepository interface {
	List(ctx context.Context, filter DriverListFilter) (*DriverListResult, error)
	Get(ctx context.Context, id int64) (*adminsvc.Driver, error)
}

type SQLDriverQueryRepository struct {
	db *sql.DB
}

func NewSQLDriverQueryRepository(db *sql.DB) DriverQueryRepository {
	return &SQLDriverQueryRepository{db: db}
}

func (r *SQLDriverQueryRepository) List(ctx context.Context, filter DriverListFilter) (*DriverListResult, error) {
	where, args := buildDriverWhere(filter)
	fromSQL := driverListFromSQL()

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) `+fromSQL+where, args...).Scan(&total); err != nil {
		return nil, err
	}

	limit := normalizePageSize(filter.PageSize)
	queryArgs := append(args, limit, offset(filter.Page, filter.PageSize))
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.phone, d.real_name, d.id_card_no, d.driver_license_no, d.avatar_url,
		       d.status, d.online_status,
		       COALESCE(v.id, 0), COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       COALESCE(c.id, 0), COALESCE(c.audit_status, 0), COALESCE(c.audit_remark, ''),
		       d.created_at, d.updated_at
	`+fromSQL+where+`
		ORDER BY d.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.Driver, 0)
	for rows.Next() {
		item, err := scanDriver(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &DriverListResult{List: list, Total: total, Page: normalizePage(filter.Page), PageSize: limit}, nil
}

func (r *SQLDriverQueryRepository) Get(ctx context.Context, id int64) (*adminsvc.Driver, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT d.id, d.phone, d.real_name, d.id_card_no, d.driver_license_no, d.avatar_url,
		       d.status, d.online_status,
		       COALESCE(v.id, 0), COALESCE(v.plate_no, ''), COALESCE(v.status, 0),
		       COALESCE(c.id, 0), COALESCE(c.audit_status, 0), COALESCE(c.audit_remark, ''),
		       d.created_at, d.updated_at
	`+driverListFromSQL()+`
		WHERE d.id = ? AND d.deleted_at IS NULL
	`, id)
	return scanDriver(row)
}

func driverListFromSQL() string {
	return `
		FROM driver d
		LEFT JOIN driver_vehicle v ON v.id = (
			SELECT MAX(v2.id) FROM driver_vehicle v2 WHERE v2.driver_id = d.id
		)
		LEFT JOIN driver_certification c ON c.id = (
			SELECT MAX(c2.id) FROM driver_certification c2 WHERE c2.driver_id = d.id
		)
	`
}

func buildDriverWhere(filter DriverListFilter) (string, []any) {
	parts := []string{"d.deleted_at IS NULL"}
	args := make([]any, 0)
	if strings.TrimSpace(filter.Keyword) != "" {
		parts = append(parts, "(d.phone LIKE ? OR d.real_name LIKE ? OR d.id_card_no LIKE ? OR d.driver_license_no LIKE ? OR v.plate_no LIKE ?)")
		kw := "%" + strings.TrimSpace(filter.Keyword) + "%"
		args = append(args, kw, kw, kw, kw, kw)
	}
	if filter.Status > 0 {
		parts = append(parts, "d.status = ?")
		args = append(args, filter.Status)
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func scanDriver(scanner interface{ Scan(...any) error }) (*adminsvc.Driver, error) {
	var item adminsvc.Driver
	var createdAt, updatedAt sql.NullTime
	err := scanner.Scan(
		&item.Id, &item.Phone, &item.RealName, &item.IdCardNo, &item.DriverLicenseNo, &item.AvatarUrl,
		&item.Status, &item.OnlineStatus, &item.VehicleId, &item.PlateNo, &item.VehicleStatus,
		&item.CertificationId, &item.AuditStatus, &item.AuditRemark, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDriverNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan driver row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

func normalizePage(page int32) int32 {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func offset(page, pageSize int32) int32 {
	return (normalizePage(page) - 1) * normalizePageSize(pageSize)
}

func formatNullTime(t sql.NullTime) string {
	if !t.Valid || t.Time.IsZero() {
		return ""
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

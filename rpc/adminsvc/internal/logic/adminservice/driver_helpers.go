package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildCertificationWhere 组装司机审核列表筛选条件。
func buildCertificationWhere(in *adminsvc.DriverCertificationListRequest) (string, []any) {
	parts := make([]string, 0)
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "(d.phone LIKE ? OR d.real_name LIKE ? OR v.plate_no LIKE ?)")
		kw := "%" + in.GetKeyword() + "%"
		args = append(args, kw, kw, kw)
	}
	if in.GetAuditStatus() > 0 {
		parts = append(parts, "c.audit_status = ?")
		args = append(args, in.GetAuditStatus())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "c.created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "c.created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

// scanCertificationRows 将司机审核列表行转换为 protobuf 对象。
func scanCertificationRows(rows *sql.Rows) (*adminsvc.DriverCertification, error) {
	var item adminsvc.DriverCertification
	var auditedAt, createdAt, updatedAt sql.NullTime
	err := rows.Scan(
		&item.Id, &item.DriverId, &item.VehicleId,
		&item.DriverPhone, &item.DriverName, &item.DriverStatus,
		&item.PlateNo, &item.VehicleStatus,
		&item.IdCardFrontUrl, &item.IdCardBackUrl, &item.DriverLicenseUrl, &item.VehicleLicenseUrl,
		&item.AuditStatus, &item.AuditRemark, &item.AuditedBy, &auditedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan certification row: %w", err)
	}
	item.AuditedAt = formatNullTime(auditedAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// auditCertification 在事务中执行司机认证审核。
func auditCertification(ctx context.Context, svcCtx *svc.ServiceContext, in *adminsvc.AuditDriverCertificationRequest, auditStatus int32) error {
	if in.GetId() <= 0 {
		return status.Error(codes.InvalidArgument, "审核记录ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	tx, err := svcCtx.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	res, err := tx.ExecContext(ctx, `
		UPDATE driver_certification
		SET audit_status = ?, audit_remark = ?, audited_by = ?, audited_at = ?, updated_at = ?
		WHERE id = ?
	`, auditStatus, in.GetRemark(), in.GetAdminId(), now, now, in.GetId())
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return status.Error(codes.NotFound, "司机审核记录不存在")
	}
	if auditStatus == 2 {
		var driverID, vehicleID int64
		if err := tx.QueryRowContext(ctx, `SELECT driver_id, vehicle_id FROM driver_certification WHERE id = ?`, in.GetId()).Scan(&driverID, &vehicleID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE driver SET status = 2, updated_at = ? WHERE id = ?`, now, driverID); err != nil {
			return err
		}
		if vehicleID > 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE driver_vehicle SET status = 2, updated_at = ? WHERE id = ?`, now, vehicleID); err != nil {
				return err
			}
		}
	}
	action := "reject"
	detail := "司机认证驳回"
	if auditStatus == 2 {
		action = "approve"
		detail = "司机认证通过"
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_operation_log (admin_id, module, action, target_type, target_id, detail, ip)
		VALUES (?, 'driver', ?, 'driver_certification', ?, ?, ?)
	`, in.GetAdminId(), action, in.GetId(), detail, in.GetIp()); err != nil {
		return err
	}
	return tx.Commit()
}

// scanCertificationRow 处理司机审核详情单行结果。
func scanCertificationRow(row *sql.Row) (*adminsvc.DriverCertification, error) {
	var item adminsvc.DriverCertification
	var auditedAt, createdAt, updatedAt sql.NullTime
	err := row.Scan(
		&item.Id, &item.DriverId, &item.VehicleId,
		&item.DriverPhone, &item.DriverName, &item.DriverStatus,
		&item.PlateNo, &item.VehicleStatus,
		&item.IdCardFrontUrl, &item.IdCardBackUrl, &item.DriverLicenseUrl, &item.VehicleLicenseUrl,
		&item.AuditStatus, &item.AuditRemark, &item.AuditedBy, &auditedAt, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "司机审核记录不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("scan certification: %w", err)
	}
	item.AuditedAt = formatNullTime(auditedAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

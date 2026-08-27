package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildCertificationWhere builds filters for the driver certification audit list.
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

// scanCertificationRows converts one driver certification list row.
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

// auditCertification routes the admin audit request to driversvc and records the audit trail.
func auditCertification(ctx context.Context, svcCtx *svc.ServiceContext, in *adminsvc.AuditDriverCertificationRequest, auditStatus int32) error {
	if in.GetId() <= 0 {
		return status.Error(codes.InvalidArgument, "audit record id cannot be empty")
	}
	if in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "admin id cannot be empty")
	}
	if svcCtx == nil || svcCtx.DriverSvc == nil {
		return status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}

	rpcReq := &driverproto.AuditCertificationRequest{
		CertificationId: in.GetId(),
		Remark:          in.GetRemark(),
		OperatorId:      in.GetAdminId(),
		Ip:              in.GetIp(),
	}
	action := "reject"
	detail := "driver certification rejected and synced to driver service"
	if auditStatus == 2 {
		if _, err := svcCtx.DriverSvc.ApproveCertification(ctx, rpcReq); err != nil {
			return err
		}
		action = "approve"
		detail = "driver certification approved and synced to driver service"
	} else {
		if _, err := svcCtx.DriverSvc.RejectCertification(ctx, rpcReq); err != nil {
			return err
		}
	}
	if err := writeAuditAfterCommitted(ctx, svcCtx, in.GetAdminId(), "driver", action, "driver_certification", in.GetId(), detail, in.GetIp()); err != nil {
		return fmt.Errorf("driver certification audit writeback failed: %w", err)
	}
	return nil
}

// scanCertificationRow converts the driver certification detail row.
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
		return nil, status.Error(codes.NotFound, "driver certification record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan certification: %w", err)
	}
	item.AuditedAt = formatNullTime(auditedAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

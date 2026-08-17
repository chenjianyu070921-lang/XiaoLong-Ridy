package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driversvcproto "XiaoLong-Ridy/rpc/driversvc/proto"

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

// auditCertification 通过 driversvc 执行司机认证审核。
// 管理后台只负责鉴权、参数转换和审计留痕，司机认证状态、司机可听单状态、车辆状态由 driversvc 在本地事务中维护。
func auditCertification(ctx context.Context, svcCtx *svc.ServiceContext, in *adminsvc.AuditDriverCertificationRequest, auditStatus int32) error {
	if in.GetId() <= 0 {
		return status.Error(codes.InvalidArgument, "审核记录ID不能为空")
	}
	if in.GetAdminId() <= 0 {
		return status.Error(codes.InvalidArgument, "管理员ID不能为空")
	}
	rpcReq := &driversvcproto.AuditCertificationRequest{
		CertificationId: in.GetId(),
		Remark:          in.GetRemark(),
		OperatorId:      in.GetAdminId(),
		Ip:              in.GetIp(),
	}
	action := "reject"
	detail := "司机认证驳回，已同步 driversvc"
	if auditStatus == 2 {
		if _, err := svcCtx.DriversSvc.ApproveCertification(ctx, rpcReq); err != nil {
			return err
		}
		action = "approve"
		detail = "司机认证通过，已同步 driversvc 并联动司机可听单状态"
	} else {
		if _, err := svcCtx.DriversSvc.RejectCertification(ctx, rpcReq); err != nil {
			return err
		}
	}
	return createOperationLog(ctx, svcCtx, in.GetAdminId(), "driver", action, "driver_certification", in.GetId(), detail, in.GetIp())
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

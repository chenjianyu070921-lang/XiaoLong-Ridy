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
	pushproto "XiaoLong-Ridy/rpc/pushesvc/pushesvc"

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

// notifyDriverBestEffort 通过 pushsvc 给司机端发送站内信和 App 推送。
// 通知属于跨服务副作用，pushsvc 未配置或调用失败时写入 admin_audit_outbox 等待异步补偿，
// 不回滚已完成的司机域状态变更。
func notifyDriverBestEffort(ctx context.Context, svcCtx *svc.ServiceContext, driverID, adminID int64, title, content, action, ip string) error {
	if svcCtx == nil || driverID <= 0 {
		return nil
	}
	if svcCtx.PushSvc == nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, "driver", action+"_notify", "driver", driverID, content, ip, status.Error(codes.FailedPrecondition, "push service is not running or downstream RPC is disabled"))
	}
	if _, err := svcCtx.PushSvc.SendNotice(ctx, &pushproto.SendNoticeReq{UserId: driverID, Title: title, Content: content, BizType: 2}); err != nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, "driver", action+"_notice", "driver", driverID, content, ip, err)
	}
	if _, err := svcCtx.PushSvc.SendPush(ctx, &pushproto.SendPushReq{UserId: driverID, Title: title, Body: content, Extras: fmt.Sprintf(`{"target_type":"driver","target_id":%d,"action":"%s"}`, driverID, action), DeviceType: "driver"}); err != nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, "driver", action+"_push", "driver", driverID, content, ip, err)
	}
	return nil
}

// freezeRiskDriverAfterBlacklist 在司机被加入黑名单后联动 driversvc 冻结账号。
// 黑名单写入成功后再调用，避免本地事务内进行跨服务 RPC；冻结或通知失败会写补偿 outbox，
// 保持黑名单主处置成功语义，后续由补偿任务重放冻结或通知。
func freezeRiskDriverAfterBlacklist(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, reason string, adminID int64, ip string) error {
	if strings.TrimSpace(reason) == "" {
		reason = "风控黑名单处置"
	}
	if svcCtx == nil {
		return status.Error(codes.FailedPrecondition, "admin service context is not initialized")
	}
	if svcCtx.DriverSvc == nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, "risk", "freeze_driver", "driver", driverID, "风控黑名单联动冻结司机："+reason, ip, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled"))
	}
	if _, err := svcCtx.DriverSvc.FreezeDriver(ctx, &driverproto.FreezeDriverRequest{
		DriverId:   driverID,
		Reason:     reason,
		OperatorId: adminID,
		Ip:         ip,
	}); err != nil {
		return recordAuditOutbox(ctx, svcCtx, adminID, "risk", "freeze_driver", "driver", driverID, "风控黑名单联动冻结司机："+reason, ip, err)
	}
	if err := writeAuditAfterCommitted(ctx, svcCtx, adminID, "risk", "freeze_driver", "driver", driverID, "风控黑名单联动冻结司机："+reason, ip); err != nil {
		return err
	}
	return notifyDriverBestEffort(ctx, svcCtx, driverID, adminID, "账号已被冻结", reason, "risk_freeze", ip)
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

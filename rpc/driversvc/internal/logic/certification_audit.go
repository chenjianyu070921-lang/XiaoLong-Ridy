package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// certificationAuditRow 承载审核所需的最小字段。
type certificationAuditRow struct {
	DriverID  uint64
	VehicleID uint64
	Status    int8
}

// loadCertificationForAudit 查询资质记录并检查当前状态。
func loadCertificationForAudit(ctx context.Context, db *gorm.DB, certificationID int64) (*certificationAuditRow, error) {
	var row certificationAuditRow
	err := db.WithContext(ctx).
		Table("driver_certification").
		Select("driver_id, vehicle_id, audit_status AS status").
		Where("id = ?", certificationID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "driver certification not found")
	}
	if err != nil {
		return nil, err
	}
	if row.Status != 1 {
		return nil, status.Error(codes.FailedPrecondition, "certification already audited")
	}
	return &row, nil
}

// updateCertificationAudit 在事务中更新审核状态与司机、车辆状态。
func updateCertificationAudit(ctx context.Context, svcCtx *svc.ServiceContext, certificationID int64, operatorID int64, remark string, auditStatus int8) error {
	if svcCtx == nil || svcCtx.DB == nil {
		return status.Error(codes.FailedPrecondition, "database not ready")
	}
	row, err := loadCertificationForAudit(ctx, svcCtx.DB, certificationID)
	if err != nil {
		return err
	}

	tx := svcCtx.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		_ = tx.Rollback().Error
	}()

	now := time.Now()
	if err := tx.Table("driver_certification").
		Where("id = ?", certificationID).
		Updates(map[string]any{
			"audit_status": auditStatus,
			"audit_remark": remark,
			"audited_by":   operatorID,
			"audited_at":   now,
			"updated_at":   now,
		}).Error; err != nil {
		return err
	}
	if auditStatus == 2 {
		if err := tx.Table("driver").Where("id = ?", row.DriverID).Updates(map[string]any{
			"status":     2,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
		if row.VehicleID > 0 {
			if err := tx.Table("driver_vehicle").Where("id = ?", row.VehicleID).Updates(map[string]any{
				"status":     2,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
	}
	return tx.Commit().Error
}

// certificationAuditResponse 返回统一成功响应。
func certificationAuditResponse() *model.DriverCertification {
	return nil
}

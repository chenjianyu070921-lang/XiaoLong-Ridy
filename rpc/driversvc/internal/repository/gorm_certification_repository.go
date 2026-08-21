package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type gormCertificationRepository struct {
	db *gorm.DB
}

// NewGormCertificationRepository 创建基于 GORM 的司机资质仓储。
func NewGormCertificationRepository(db *gorm.DB) CertificationRepository {
	return &gormCertificationRepository{db: db}
}

// Upsert 写入/更新司机资质记录：按 driver_id 幂等 upsert，返回最新记录。
func (r *gormCertificationRepository) Upsert(ctx context.Context, cert *model.DriverCertification) (*model.DriverCertification, error) {
	// 按 driver_id 做原子 upsert：存在则更新非空字段，不存在则插入。
	if err := r.db.WithContext(ctx).
		Where("driver_id = ?", cert.DriverId).
		Assign(*cert).
		FirstOrCreate(cert).Error; err != nil {
		return nil, err
	}
	return cert, nil
}

// GetByDriverID 按司机 ID 查询资质记录；未找到返回 (nil, ErrCertificationNotFound)。
func (r *gormCertificationRepository) GetByDriverID(ctx context.Context, driverID uint64) (*model.DriverCertification, error) {
	var cert model.DriverCertification
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&cert).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, ErrCertificationNotFound
		}
		return nil, err
	}
	return &cert, nil
}

type certificationAuditRow struct {
	DriverID  uint64
	VehicleID uint64
	Status    int8
}

// UpdateAudit 更新司机资质审核状态，并在通过时联动司机与车辆状态。
func (r *gormCertificationRepository) UpdateAudit(ctx context.Context, certificationID int64, operatorID int64, remark string, auditStatus int8) error {
	if certificationID <= 0 {
		return status.Error(codes.InvalidArgument, "certification id is required")
	}
	if operatorID <= 0 {
		return status.Error(codes.InvalidArgument, "operator id is required")
	}
	if auditStatus == 3 && strings.TrimSpace(remark) == "" {
		return status.Error(codes.InvalidArgument, "reject remark is required")
	}

	var row certificationAuditRow
	err := r.db.WithContext(ctx).
		Table("driver_certification").
		Select("driver_id, vehicle_id, audit_status AS status").
		Where("id = ?", certificationID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "driver certification not found")
	}
	if err != nil {
		return err
	}
	if row.Status != 1 {
		return status.Error(codes.FailedPrecondition, "certification already audited")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		_ = tx.Rollback().Error
	}()

	now := time.Now()
	if err := tx.Table("driver_certification").Where("id = ?", certificationID).Updates(map[string]any{
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

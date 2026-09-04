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

// adminCertificationSelectSQL 是认证记录与司机、车辆摘要的关联查询字段（供管理后台审核页展示）。
const adminCertificationSelectSQL = `
	SELECT c.id AS id, c.driver_id AS driver_id, c.vehicle_id AS vehicle_id,
	       COALESCE(d.phone, '') AS driver_phone, COALESCE(d.real_name, '') AS driver_name,
	       COALESCE(d.status, 0) AS driver_status,
	       COALESCE(v.plate_no, '') AS plate_no, COALESCE(v.status, 0) AS vehicle_status,
	       c.id_card_front_url AS id_card_front_url, c.id_card_back_url AS id_card_back_url,
	       c.driver_license_url AS driver_license_url, c.vehicle_license_url AS vehicle_license_url,
	       c.audit_status AS audit_status, c.audit_remark AS audit_remark, c.audited_by AS audited_by,
	       c.audited_at AS audited_at, c.created_at AS created_at, c.updated_at AS updated_at
	FROM driver_certification c
	LEFT JOIN driver d ON d.id = c.driver_id
	LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id`

// buildAdminCertificationWhere 组装后台认证审核列表筛选条件。
func buildAdminCertificationWhere(filter AdminCertificationFilter) (string, []any) {
	parts := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		parts = append(parts, "(d.phone LIKE ? OR d.real_name LIKE ? OR v.plate_no LIKE ?)")
		like := "%" + kw + "%"
		args = append(args, like, like, like)
	}
	if filter.AuditStatus > 0 {
		parts = append(parts, "c.audit_status = ?")
		args = append(args, filter.AuditStatus)
	}
	if !filter.StartTime.IsZero() {
		parts = append(parts, "c.created_at >= ?")
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		parts = append(parts, "c.created_at <= ?")
		args = append(args, filter.EndTime)
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

// AdminList 管理后台分页查询认证记录（关联司机/车辆摘要）。
func (r *gormCertificationRepository) AdminList(ctx context.Context, filter AdminCertificationFilter) ([]*AdminCertificationRow, int64, error) {
	page, pageSize := filter.Page, filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	where, args := buildAdminCertificationWhere(filter)
	countFrom := "FROM driver_certification c LEFT JOIN driver d ON d.id = c.driver_id LEFT JOIN driver_vehicle v ON v.id = c.vehicle_id"

	var total int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(1) "+countFrom+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]*AdminCertificationRow, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err := r.db.WithContext(ctx).Raw(adminCertificationSelectSQL+where+" ORDER BY c.id DESC LIMIT ? OFFSET ?", queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// AdminGetByID 管理后台按认证记录 ID 查询详情（关联司机/车辆摘要）。
func (r *gormCertificationRepository) AdminGetByID(ctx context.Context, id uint64) (*AdminCertificationRow, error) {
	var row AdminCertificationRow
	if err := r.db.WithContext(ctx).Raw(adminCertificationSelectSQL+" WHERE c.id = ?", id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, ErrCertificationNotFound
	}
	return &row, nil
}

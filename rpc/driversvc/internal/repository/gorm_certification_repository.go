package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

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

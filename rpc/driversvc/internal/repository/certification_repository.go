package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// ErrCertificationNotFound 表示未找到指定的司机资质记录。
var ErrCertificationNotFound = errors.New("certification not found")

// CertificationRepository 定义司机资质数据访问接口，使 logic 层与具体存储实现解耦。
type CertificationRepository interface {
	// Upsert 写入/更新司机资质记录：按 driver_id 幂等，存在则更新、不存在则插入，返回最新记录。
	Upsert(ctx context.Context, cert *model.DriverCertification) (*model.DriverCertification, error)
	// GetByDriverID 按司机 ID 查询资质记录；未找到返回 (nil, ErrCertificationNotFound)。
	GetByDriverID(ctx context.Context, driverID uint64) (*model.DriverCertification, error)
}

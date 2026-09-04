package repository

import (
	"context"
	"errors"
	"time"

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
	// UpdateAudit 更新资质审核状态，并在通过时联动司机与车辆状态。
	UpdateAudit(ctx context.Context, certificationID int64, operatorID int64, remark string, auditStatus int8) error
	// AdminList 管理后台分页查询认证记录（关联司机/车辆摘要），供 adminsvc 消费，避免后台直连司机域表。
	AdminList(ctx context.Context, filter AdminCertificationFilter) ([]*AdminCertificationRow, int64, error)
	// AdminGetByID 管理后台按认证记录 ID 查询详情（关联司机/车辆摘要）。
	AdminGetByID(ctx context.Context, id uint64) (*AdminCertificationRow, error)
}

// AdminCertificationFilter 表示管理后台认证审核列表筛选条件。
type AdminCertificationFilter struct {
	Page        int
	PageSize    int
	Keyword     string
	AuditStatus int8
	StartTime   time.Time
	EndTime     time.Time
}

// AdminCertificationRow 是认证记录与司机、车辆摘要的关联查询结果，供管理后台展示。
type AdminCertificationRow struct {
	Id                uint64
	DriverId          uint64
	VehicleId         uint64
	DriverPhone       string
	DriverName        string
	DriverStatus      int8
	PlateNo           string
	VehicleStatus     int8
	IdCardFrontUrl    string
	IdCardBackUrl     string
	DriverLicenseUrl  string
	VehicleLicenseUrl string
	AuditStatus       int8
	AuditRemark       string
	AuditedBy         uint64
	AuditedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

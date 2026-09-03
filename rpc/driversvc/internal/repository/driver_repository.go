package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// ErrDriverNotFound 表示未找到指定的司机记录（含已软删记录）。
var ErrDriverNotFound = errors.New("driver not found")

// DriverRepository 定义司机数据访问接口，使 logic 层与具体存储实现解耦。
type DriverRepository interface {
	// Create 写入一条司机记录。
	Create(ctx context.Context, driver *model.Driver) error
	// GetByID 按主键查询司机（软删记录不可见）。
	GetByID(ctx context.Context, id uint64) (*model.Driver, error)
	// GetByPhone 按手机号查询司机（软删记录不可见），用于登录场景。
	GetByPhone(ctx context.Context, phone string) (*model.Driver, error)
	// List 分页查询司机列表，支持状态与关键字过滤，返回本页记录与符合条件的总数。
	List(ctx context.Context, filter DriverListFilter) ([]*model.Driver, int64, error)
	// ListNearbyDrivers 按经纬度 + 半径查找在线司机（同库查 driver_location）。
	ListNearbyDrivers(ctx context.Context, filter NearbyDriverFilter) ([]*model.DriverLocation, error)
	// UpsertLocation writes or updates the driver's latest location.
	UpsertLocation(ctx context.Context, location *model.DriverLocation) error
	// UpdateLocationStatus updates online_status in driver_location.
	UpdateLocationStatus(ctx context.Context, driverID uint64, status int8) error
	// UpdateStatusAndLocation 在一个事务里同时更新 driver 表和 driver_location 表的 online_status。
	UpdateStatusAndLocation(ctx context.Context, driverID uint64, status int8) error
	// GetDriverScore returns nil, nil when no score record exists.
	GetDriverScore(ctx context.Context, driverID uint64) (*model.DriverScore, error)
	// RefreshDriverScoreMetrics recalculates this month's score factors from source tables and upserts driver_score.
	RefreshDriverScoreMetrics(ctx context.Context, driverID uint64, startAt, endAt time.Time) (*model.DriverScore, error)
	// Update 按 ID 增量更新司机字段。
	Update(ctx context.Context, id uint64, updates map[string]interface{}) error
	// Delete 软删除指定司机（设置 deleted_at）。
	Delete(ctx context.Context, driver *model.Driver) error
}

// NearbyDriverFilter 附近司机查询过滤条件。
type NearbyDriverFilter struct {
	// Longitude 查询中心经度。
	Longitude float64
	// Latitude 查询中心纬度。
	Latitude float64
	// RadiusMeters 搜索半径（米）。
	RadiusMeters float64
	// Limit 返回条数上限。
	Limit int
}

// DriverStatusFilter 司机状态过滤条件，nil 表示不过滤。
type DriverStatusFilter = int8

// DriverListFilter 司机列表查询过滤条件。
type DriverListFilter struct {
	// Page 页码（从 1 开始）。
	Page int32
	// PageSize 每页条数。
	PageSize int32
	// Status 可选状态过滤；空指针表示不过滤。
	Status *DriverStatusFilter
	// Keyword 可选关键字，模糊匹配手机号/姓名/车牌号。
	Keyword string
}

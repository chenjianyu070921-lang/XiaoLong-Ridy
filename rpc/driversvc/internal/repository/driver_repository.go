package repository

import (
	"context"
	"errors"

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
	// UpsertLocation 写入/更新司机最新位置与在线状态（driver_location 表，按 driver_id 幂等 upsert）。
	UpsertLocation(ctx context.Context, loc *model.DriverLocation) error
	// UpdateLocationStatus 仅更新司机位置表中的在线状态，不覆盖最新经纬度。
	UpdateLocationStatus(ctx context.Context, driverID uint64, onlineStatus int8) error
	// Update 按 ID 增量更新司机字段。
	Update(ctx context.Context, id uint64, updates map[string]interface{}) error
	// Delete 软删除指定司机（设置 deleted_at）。
	Delete(ctx context.Context, driver *model.Driver) error
	// GetDriverScore 按司机 ID 查询其服务分与运营指标（driver_score 表）。
	GetDriverScore(ctx context.Context, driverID uint64) (*model.DriverScore, error)
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

package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

// errorsIsNotFound 判断是否为 GORM 记录不存在错误。
func errorsIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

type gormDriverRepository struct {
	db *gorm.DB
}

// NewGormDriverRepository 创建基于 GORM 的司机仓储。
func NewGormDriverRepository(db *gorm.DB) DriverRepository {
	return &gormDriverRepository{db: db}
}

// Create 写入一条司机记录。
func (r *gormDriverRepository) Create(ctx context.Context, driver *model.Driver) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

// GetByID 按主键查询司机，软删记录不可见。
func (r *gormDriverRepository) GetByID(ctx context.Context, id uint64) (*model.Driver, error) {
	var driver model.Driver
	err := r.db.WithContext(ctx).First(&driver, id).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}
	return &driver, nil
}

// GetByPhone 按手机号查询司机（软删记录不可见），用于登录场景。
func (r *gormDriverRepository) GetByPhone(ctx context.Context, phone string) (*model.Driver, error) {
	var driver model.Driver
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&driver).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}
	return &driver, nil
}

// Update 按 ID 增量更新司机字段。
func (r *gormDriverRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Driver{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 软删除指定司机（GORM 自动设置 deleted_at）。
func (r *gormDriverRepository) Delete(ctx context.Context, driver *model.Driver) error {
	return r.db.WithContext(ctx).Delete(driver).Error
}

// List 分页查询司机列表，支持状态与关键字（手机号/姓名）过滤。
// 返回本页记录与符合条件的总记录数；软删记录不可见。
func (r *gormDriverRepository) List(ctx context.Context, filter DriverListFilter) ([]*model.Driver, int64, error) {
	// 构造带过滤条件的查询作用域。
	query := r.db.WithContext(ctx).Model(&model.Driver{})
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Keyword != "" {
		// 模糊匹配手机号或姓名。
		like := "%" + filter.Keyword + "%"
		query = query.Where("phone LIKE ? OR real_name LIKE ?", like, like)
	}

	// 统计符合条件的总记录数。
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页参数收敛：页码至少 1，每页至少 1 条、至多 100 条。
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询本页数据（按 ID 倒序，新注册的靠前）。
	var drivers []*model.Driver
	if err := query.Order("id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&drivers).Error; err != nil {
		return nil, 0, err
	}
	return drivers, total, nil
}

// earthRadiusMeters 地球平均半径（米），用于 Haversine 球面距离计算。
const earthRadiusMeters = 6371000.0

// ListNearbyDrivers 按经纬度 + 半径查找在线司机（online_status=1）。
// 使用 Haversine 公式在 SQL 中直接计算球面距离，过滤半径内记录并按距离升序返回。
func (r *gormDriverRepository) ListNearbyDrivers(ctx context.Context, filter NearbyDriverFilter) ([]*model.DriverLocation, error) {
	// 半径收敛：缺省 3000 米，单程派单场景上限 50000 米。
	radius := filter.RadiusMeters
	if radius <= 0 {
		radius = 3000
	}
	// 条数收敛：缺省 20，上限 100。
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Haversine 公式：计算每条记录到中心点的球面距离（米）。
	// 6371000 为地球半径；经度/纬度需转为弧度参与计算。
	// 用 Raw 手写 SQL，参数顺序固定为：
	//   [SELECT 中 haversine] 中心点经度、中心点纬度、中心点纬度
	//   online_status 过滤值
	//   [WHERE 中 haversine] 中心点经度、中心点纬度、中心点纬度
	//   半径、limit
	haversine := `6371000 * 2 * ASIN(SQRT(
		POWER(SIN(RADIANS(longitude - ?) / 2), 2) +
		COS(RADIANS(?)) * COS(RADIANS(latitude)) * POWER(SIN(RADIANS(latitude - ?) / 2), 2)
	))`

	sql := `SELECT *, (` + haversine + `) AS distance_meters
		FROM driver_location
		WHERE online_status = ?
		  AND (` + haversine + `) <= ?
		ORDER BY distance_meters ASC
		LIMIT ?`

	var locations []*model.DriverLocation
	if err := r.db.WithContext(ctx).
		Raw(sql,
			filter.Longitude, filter.Latitude, filter.Latitude, // SELECT haversine
			model.LocationOnline,                  // online_status 过滤
			filter.Longitude, filter.Latitude, filter.Latitude, // WHERE haversine
			radius,  // 半径
			limit,   // 条数
		).Scan(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

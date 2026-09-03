package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm/clause"
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
		query = query.Where("driver.status = ?", *filter.Status)
	}
	if filter.Keyword != "" {
		// 模糊匹配司机身份字段和车牌号，保持后台司机查询的搜索语义。
		like := "%" + filter.Keyword + "%"
		query = query.Joins("LEFT JOIN driver_vehicle v ON v.driver_id = driver.id").
			Where("driver.phone LIKE ? OR driver.real_name LIKE ? OR driver.id_card_no LIKE ? OR driver.driver_license_no LIKE ? OR v.plate_no LIKE ?", like, like, like, like, like)
	}

	// 统计符合条件的总记录数。
	var total int64
	countQuery := query
	if filter.Keyword != "" {
		countQuery = countQuery.Distinct("driver.id")
	}
	if err := countQuery.Count(&total).Error; err != nil {
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
	if filter.Keyword != "" {
		query = query.Select("driver.*").Group("driver.id")
	}
	if err := query.Order("driver.id DESC").
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
			model.LocationOnline,                               // online_status 过滤
			filter.Longitude, filter.Latitude, filter.Latitude, // WHERE haversine
			radius, // 半径
			limit,  // 条数
		).Scan(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

// UpsertLocation writes the latest driver location, creating the row when needed.
func (r *gormDriverRepository) UpsertLocation(ctx context.Context, location *model.DriverLocation) error {
	if location == nil {
		return nil
	}

	updates := map[string]interface{}{
		"longitude":     location.Longitude,
		"latitude":      location.Latitude,
		"heading":       location.Heading,
		"speed_kmh":     location.SpeedKmh,
		"online_status": location.OnlineStatus,
		"report_time":   location.ReportTime,
	}
	// 使用数据库原子 Upsert，避免“先 UPDATE 再 INSERT”在并发心跳下触发唯一键冲突。
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "driver_id"}},
		DoUpdates: clause.Assignments(updates),
	}).Model(&model.DriverLocation{}).Create(map[string]interface{}{
		"driver_id":     location.DriverID,
		"longitude":     location.Longitude,
		"latitude":      location.Latitude,
		"heading":       location.Heading,
		"speed_kmh":     location.SpeedKmh,
		"online_status": location.OnlineStatus,
		"report_time":   location.ReportTime,
	}).Error
}

// UpdateLocationStatus updates the cached location row status for a driver.
func (r *gormDriverRepository) UpdateLocationStatus(ctx context.Context, driverID uint64, status int8) error {
	return r.db.WithContext(ctx).
		Model(&model.DriverLocation{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"online_status": status,
			"report_time":   time.Now(),
		}).Error
}

// UpdateStatusAndLocation 在一个事务里同时更新 driver 表和 driver_location 表的 online_status，避免中间状态不一致。
func (r *gormDriverRepository) UpdateStatusAndLocation(ctx context.Context, driverID uint64, status int8) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Driver{}).
			Where("id = ?", driverID).
			Updates(map[string]interface{}{"online_status": status}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.DriverLocation{}).
			Where("driver_id = ?", driverID).
			Updates(map[string]interface{}{
				"online_status": status,
				"report_time":   time.Now(),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetDriverScore returns the driver's scoring metrics.
func (r *gormDriverRepository) GetDriverScore(ctx context.Context, driverID uint64) (*model.DriverScore, error) {
	var score model.DriverScore
	err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		First(&score).Error
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &score, nil
}

func (r *gormDriverRepository) RefreshDriverScoreMetrics(ctx context.Context, driverID uint64, startAt, endAt time.Time) (*model.DriverScore, error) {
	var completedOrders int64
	if err := r.db.WithContext(ctx).Table("ride_order").
		Where("driver_id = ? AND status = ? AND updated_at >= ? AND updated_at < ?", driverID, constants.OrderStatusCompleted, startAt, endAt).
		Count(&completedOrders).Error; err != nil {
		return nil, err
	}

	var cancelledOrders int64
	if err := r.db.WithContext(ctx).Table("ride_order").
		Where("driver_id = ? AND status = ? AND updated_at >= ? AND updated_at < ?", driverID, constants.OrderStatusCancelled, startAt, endAt).
		Count(&cancelledOrders).Error; err != nil {
		return nil, err
	}

	var avgRating sql.NullFloat64
	if err := r.db.WithContext(ctx).Table("order_review").
		Select("AVG(rating)").
		Where("driver_id = ? AND created_at >= ? AND created_at < ?", driverID, startAt, endAt).
		Scan(&avgRating).Error; err != nil {
		return nil, err
	}

	var complaintCount int64
	if err := r.db.WithContext(ctx).Table("admin_complaint_work_order").
		Where("driver_id = ? AND created_at >= ? AND created_at < ?", driverID, startAt, endAt).
		Count(&complaintCount).Error; err != nil {
		return nil, err
	}

	existing, err := r.GetDriverScore(ctx, driverID)
	if err != nil {
		return nil, err
	}

	serviceScore := 100.0
	if existing != nil && existing.Score > 0 {
		serviceScore = existing.Score
	}
	if avgRating.Valid {
		serviceScore = avgRating.Float64 * 20.0
	}
	totalFinishedOrCancelled := completedOrders + cancelledOrders
	cancelRate := 0.0
	if totalFinishedOrCancelled > 0 {
		cancelRate = float64(cancelledOrders) * 100.0 / float64(totalFinishedOrCancelled)
	}

	now := time.Now()
	score := &model.DriverScore{
		DriverId:            driverID,
		Score:               serviceScore,
		Level:               driverLevelFromServiceScore(serviceScore),
		MonthOrders:         int(completedOrders),
		MonthCancelRate:     cancelRate,
		MonthComplaintCount: int(complaintCount),
		UpdatedAt:           now,
	}
	if existing == nil {
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "driver_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"score",
				"level",
				"month_orders",
				"month_cancel_rate",
				"month_complaint_count",
				"updated_at",
			}),
		}).Create(score).Error; err != nil {
			return nil, err
		}
		return score, nil
	}
	score.Id = existing.Id
	if err := r.db.WithContext(ctx).Model(&model.DriverScore{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"score":                 score.Score,
			"level":                 score.Level,
			"month_orders":          score.MonthOrders,
			"month_cancel_rate":     score.MonthCancelRate,
			"month_complaint_count": score.MonthComplaintCount,
			"updated_at":            now,
		}).Error; err != nil {
		return nil, err
	}
	return score, nil
}

func driverLevelFromServiceScore(score float64) int8 {
	switch {
	case score >= 95:
		return 5
	case score >= 90:
		return 4
	case score >= 80:
		return 3
	case score >= 70:
		return 2
	default:
		return 1
	}
}

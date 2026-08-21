package repository

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/pricesvc/internal/model"

	"gorm.io/gorm"
)

// PriceRuleRepo 计价规则数据访问。
type PriceRuleRepo struct {
	db *gorm.DB
}

func NewPriceRuleRepo(db *gorm.DB) *PriceRuleRepo {
	return &PriceRuleRepo{db: db}
}

// FindActive 查询启用的计价规则：优先城市精确匹配，其次全局（city_code 为空）。
func (r *PriceRuleRepo) FindActive(ctx context.Context, cityCode string, carType int8) (*model.PriceRule, error) {
	var rule model.PriceRule
	err := r.db.WithContext(ctx).
		Where("status = ? AND car_type = ? AND (city_code = ? OR city_code = '')", 1, carType, cityCode).
		Order("city_code DESC"). // 城市精确匹配优先于全局（非空 city_code 排前）
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// List 查询计价规则列表并返回分页总数。
func (r *PriceRuleRepo) List(ctx context.Context, keyword, cityCode string, carType, status int32, limit, offset int) ([]model.PriceRule, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.PriceRule{})
	if keyword != "" {
		like := "%" + strings.TrimSpace(keyword) + "%"
		db = db.Where("name LIKE ? OR city_code LIKE ?", like, like)
	}
	if cityCode != "" {
		db = db.Where("city_code = ?", cityCode)
	}
	if carType > 0 {
		db = db.Where("car_type = ?", carType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PriceRule
	if err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetByID 按 ID 查询计价规则详情。
func (r *PriceRuleRepo) GetByID(ctx context.Context, id int64) (*model.PriceRule, error) {
	var rule model.PriceRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create 新增计价规则。
func (r *PriceRuleRepo) Create(ctx context.Context, rule *model.PriceRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// Update 更新计价规则。
func (r *PriceRuleRepo) Update(ctx context.Context, rule *model.PriceRule) error {
	res := r.db.WithContext(ctx).
		Model(&model.PriceRule{}).
		Where("id = ?", rule.Id).
		Updates(map[string]any{
			"name":               rule.Name,
			"city_code":          rule.CityCode,
			"car_type":           rule.CarType,
			"base_price":         rule.BasePrice,
			"base_distance_km":   rule.BaseDistanceKm,
			"per_km_price":       rule.PerKmPrice,
			"per_minute_price":   rule.PerMinutePrice,
			"night_start_time":   rule.NightStartTime,
			"night_end_time":     rule.NightEndTime,
			"night_surcharge":    rule.NightSurcharge,
			"dynamic_max_factor": rule.DynamicMaxFactor,
			"status":             rule.Status,
			"effective_at":       rule.EffectiveAt,
			"expire_at":          rule.ExpireAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateStatus 更新计价规则状态。
func (r *PriceRuleRepo) UpdateStatus(ctx context.Context, id int64, status int32) error {
	res := r.db.WithContext(ctx).
		Model(&model.PriceRule{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

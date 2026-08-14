package repository

import (
	"context"

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

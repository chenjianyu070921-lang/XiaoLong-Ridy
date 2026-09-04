package repository

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"

	"gorm.io/gorm"
)

// SettlementRepo 结算单数据访问。
type SettlementRepo struct {
	db *gorm.DB
}

func NewSettlementRepo(db *gorm.DB) *SettlementRepo {
	return &SettlementRepo{db: db}
}

// Create 创建结算单。
func (r *SettlementRepo) Create(ctx context.Context, s *model.Settlement) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// FindByOrderId 按订单ID查询结算单。
func (r *SettlementRepo) FindByOrderId(ctx context.Context, orderId uint64) (*model.Settlement, error) {
	var s model.Settlement
	err := r.db.WithContext(ctx).Where("order_id = ?", orderId).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type SettlementListFilter struct {
	DriverID uint64
	Status   int32
	StartAt  *time.Time
	EndAt    *time.Time
	Page     int32
	PageSize int32
}

func (r *SettlementRepo) List(ctx context.Context, filter SettlementListFilter) ([]*model.Settlement, int64, error) {
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

	query := r.db.WithContext(ctx).Model(&model.Settlement{})
	if filter.DriverID > 0 {
		query = query.Where("driver_id = ?", filter.DriverID)
	}
	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartAt != nil {
		query = query.Where("settled_at >= ?", *filter.StartAt)
	}
	if filter.EndAt != nil {
		query = query.Where("settled_at < ?", *filter.EndAt)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []*model.Settlement
	if err := query.Order("settled_at DESC,id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

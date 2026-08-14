package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/pricesvc/internal/model"

	"gorm.io/gorm"
)

// OrderPriceRepo 订单价格明细数据访问。
type OrderPriceRepo struct {
	db *gorm.DB
}

func NewOrderPriceRepo(db *gorm.DB) *OrderPriceRepo {
	return &OrderPriceRepo{db: db}
}

// Create 插入一条价格明细。
func (r *OrderPriceRepo) Create(ctx context.Context, p *model.OrderPrice) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// FindByOrderId 按订单ID查询。
func (r *OrderPriceRepo) FindByOrderId(ctx context.Context, orderId uint64) (*model.OrderPrice, error) {
	var p model.OrderPrice
	err := r.db.WithContext(ctx).Where("order_id = ?", orderId).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update 更新价格明细。
func (r *OrderPriceRepo) Update(ctx context.Context, p *model.OrderPrice) error {
	return r.db.WithContext(ctx).Save(p).Error
}

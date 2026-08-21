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

// UpdateSelective 按 id 条件更新指定列，避免 Save 全字段覆盖：
//   1) 并发场景下丢字段；
//   2) created_at 被清零或被无关值污染；
//   3) 引用零值 decimal 字段被覆写。
// 强烈建议在事务内调用。
func (r *OrderPriceRepo) UpdateSelective(ctx context.Context, id uint64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.OrderPrice{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Update 保留旧接口（向后兼容），新代码请使用 UpdateSelective。
func (r *OrderPriceRepo) Update(ctx context.Context, p *model.OrderPrice) error {
	return r.db.WithContext(ctx).Save(p).Error
}

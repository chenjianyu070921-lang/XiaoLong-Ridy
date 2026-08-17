package repository

import (
	"context"

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

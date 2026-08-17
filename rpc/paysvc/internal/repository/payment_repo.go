package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"

	"gorm.io/gorm"
)

// PaymentRepo 支付单数据访问。
type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// Create 创建支付单。
func (r *PaymentRepo) Create(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// FindByPaymentNo 按支付单号查询。
func (r *PaymentRepo) FindByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error) {
	var p model.Payment
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindByOrderId 按订单ID查询最新一条支付单。
func (r *PaymentRepo) FindByOrderId(ctx context.Context, orderId uint64) (*model.Payment, error) {
	var p model.Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderId).
		Order("id DESC").
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update 更新支付单。
func (r *PaymentRepo) Update(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

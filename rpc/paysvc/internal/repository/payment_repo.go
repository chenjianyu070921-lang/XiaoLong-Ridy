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

// FindUnsentPaidPayments 查询已支付但 Kafka 事件未发送的支付单，用于对账补发。
func (r *PaymentRepo) FindUnsentPaidPayments(ctx context.Context, limit int) ([]*model.Payment, error) {
	var list []*model.Payment
	if err := r.db.WithContext(ctx).
		Where("status = ? AND event_sent = ?", model.PaymentStatusPaid, false).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindUnsettledPaidPayments 拉取已支付但未结算（settlement 表无对应订单记录）的支付单，供自动结算 job 使用。
func (r *PaymentRepo) FindUnsettledPaidPayments(ctx context.Context, limit int) ([]*model.Payment, error) {
	var list []*model.Payment
	err := r.db.WithContext(ctx).
		Table("payment AS p").
		Joins("LEFT JOIN settlement AS s ON s.order_id = p.order_id").
		Where("p.status = ? AND s.id IS NULL", model.PaymentStatusPaid).
		Order("p.id ASC").
		Limit(limit).
		Scan(&list).Error
	return list, err
}

// UpdateSelective 按 id 条件更新指定列（不在事务外单独使用时应放在事务里）。
// 用 Updates(map) 仅更新给定列，避免 Save 全字段覆盖造成的：
//   1) 并发场景下丢字段；
//   2) created_at 被清零；
//   3) decimals/空值污染。
func (r *PaymentRepo) UpdateSelective(ctx context.Context, id uint64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Update 是向后兼容保留的"全量 Save"，已不建议在新代码里使用。
// 显式保留是因为 model.Payment 上的关联字段（如 PaidAt）零值会被写成 NULL，符合早期测试期望。
func (r *PaymentRepo) Update(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

package svc

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrReviewAlreadyExists = errors.New("review already exists")

// OrderReview 表示乘客对已完成订单的评价记录。
type OrderReview struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	OrderID   uint64    `gorm:"column:order_id"`
	UserID    uint64    `gorm:"column:user_id"`
	DriverID  uint64    `gorm:"column:driver_id"`
	Rating    int8      `gorm:"column:rating"`
	Comment   string    `gorm:"column:comment"`
	Tags      string    `gorm:"column:tags"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 返回评价表名。
func (OrderReview) TableName() string {
	return "order_review"
}

// ReviewRepository 定义评价写入仓储契约。
type ReviewRepository interface {
	Create(ctx context.Context, review *OrderReview) error
}

// GormReviewRepository 是 MySQL/GORM 评价仓储。
type GormReviewRepository struct {
	db *gorm.DB
}

// NewGormReviewRepository 创建生产环境评价仓储。
func NewGormReviewRepository(db *gorm.DB) *GormReviewRepository {
	return &GormReviewRepository{db: db}
}

// Create 写入评价，依赖 order_id 唯一索引防止重复评价。
func (r *GormReviewRepository) Create(ctx context.Context, review *OrderReview) error {
	err := r.db.WithContext(ctx).Create(review).Error
	if err != nil && isDuplicateReview(err) {
		return ErrReviewAlreadyExists
	}
	return err
}

// isDuplicateReview 判断 MySQL 唯一键冲突。
func isDuplicateReview(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || containsDuplicateKey(err.Error()))
}

// containsDuplicateKey 兼容未启用 TranslateError 时的 MySQL 重复键错误文本。
func containsDuplicateKey(message string) bool {
	return strings.Contains(message, "Duplicate entry") || strings.Contains(message, "1062")
}

package svc

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrReviewAlreadyExists = errors.New("review already exists")

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

func (OrderReview) TableName() string { return "order_review" }

type ReviewRepository interface {
	Create(ctx context.Context, review *OrderReview) error
}

type GormReviewRepository struct{ db *gorm.DB }

func NewGormReviewRepository(db *gorm.DB) *GormReviewRepository {
	return &GormReviewRepository{db: db}
}

func (r *GormReviewRepository) Create(ctx context.Context, review *OrderReview) error {
	err := r.db.WithContext(ctx).Create(review).Error
	if err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062")) {
		return ErrReviewAlreadyExists
	}
	return err
}

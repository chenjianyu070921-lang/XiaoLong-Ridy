package driver

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PassengerReview 乘客评价储存表（passenger_review）的一条原始评价记录。
// 直接接受乘客评价落库，作为自动评价 Agent 打分的乘客文本数据源。
type PassengerReview struct {
	ID             uint64    `gorm:"primaryKey;column:id"`
	DriverPhone    string    `gorm:"column:driver_phone"`    // 司机手机号（Agent 查询键）
	OrderID        string    `gorm:"column:order_id"`        // 关联订单号（可为空串）
	PassengerPhone string    `gorm:"column:passenger_phone"` // 乘客手机号（可脱敏存储）
	Rating         int       `gorm:"column:rating"`          // 乘客星级 0-5；0 表示纯文字无星级
	Comment        string    `gorm:"column:comment"`         // 乘客文字评价（打分的主要依据）
	Tags           string    `gorm:"column:tags"`            // 乘客勾选标签，逗号分隔（原始留存）
	CreatedAt      time.Time `gorm:"column:created_at"`
}

// TableName 返回乘客评价储存表名。
func (PassengerReview) TableName() string {
	return "passenger_review"
}

// ReviewStore 乘客评价存取契约：写入原始评价 + 按司机手机号查询评价。
type ReviewStore interface {
	// Save 保存一条乘客评价。
	Save(ctx context.Context, review *PassengerReview) error
	// ListByDriverPhone 按司机手机号查询最近的乘客评价（created_at 倒序，最多 limit 条）。
	ListByDriverPhone(ctx context.Context, driverPhone string, limit int) ([]PassengerReview, error)
}

// GormReviewStore 基于 GORM/MySQL 的乘客评价仓储实现。
type GormReviewStore struct {
	db *gorm.DB
}

// NewGormReviewStore 创建乘客评价仓储；db 为空返回错误。
func NewGormReviewStore(db *gorm.DB) (*GormReviewStore, error) {
	if db == nil {
		return nil, errors.New("gorm db required")
	}
	return &GormReviewStore{db: db}, nil
}

// Save 保存乘客评价；司机手机号必填，星级范围 0-5。
func (s *GormReviewStore) Save(ctx context.Context, review *PassengerReview) error {
	if review == nil {
		return errors.New("review required")
	}
	if strings.TrimSpace(review.DriverPhone) == "" {
		return errors.New("driver_phone required")
	}
	if review.Rating < 0 || review.Rating > 5 {
		return errors.New("rating out of range")
	}
	return s.db.WithContext(ctx).Create(review).Error
}

// ListByDriverPhone 查询指定司机手机号最近的乘客评价。
func (s *GormReviewStore) ListByDriverPhone(ctx context.Context, driverPhone string, limit int) ([]PassengerReview, error) {
	driverPhone = strings.TrimSpace(driverPhone)
	if driverPhone == "" {
		return nil, errors.New("driver_phone required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var list []PassengerReview
	err := s.db.WithContext(ctx).
		Where("driver_phone = ?", driverPhone).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

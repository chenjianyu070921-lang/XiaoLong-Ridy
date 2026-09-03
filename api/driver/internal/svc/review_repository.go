package svc

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ErrReviewAlreadyExists 表示同一订单已经评价过。
var ErrReviewAlreadyExists = errors.New("review already exists")

// PassengerReview 表示乘客对司机的评价，直接映射乘客端 order_review 表（只读）。
type PassengerReview struct {
	ID        uint64 `gorm:"primaryKey;column:id"`
	OrderID   uint64 `gorm:"column:order_id"`
	UserID    uint64 `gorm:"column:user_id"`
	DriverID  uint64 `gorm:"column:driver_id"`
	Rating    int8   `gorm:"column:rating"`
	Comment   string `gorm:"column:comment"`
	Tags      string `gorm:"column:tags"`
	CreatedAt int64  `gorm:"column:created_at"`
}

// TableName 返回乘客评价表名。
func (PassengerReview) TableName() string {
	return "order_review"
}

// DriverOrderReview 表示司机对乘客的评价，映射司机端 driver_review 表。
type DriverOrderReview struct {
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

// TableName 返回司机评价表名。
func (DriverOrderReview) TableName() string {
	return "driver_review"
}

// ReviewRepository 定义司机端评价的查询与写入契约。
type ReviewRepository interface {
	ListPassengerReviewsByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]PassengerReview, int64, error)
	CreateDriverReview(ctx context.Context, review *DriverOrderReview) error
	ListDriverReviewsByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]DriverOrderReview, int64, error)
}

// GormDriverReviewRepository 是司机端评价的 MySQL/GORM 仓储实现。
type GormDriverReviewRepository struct {
	db *gorm.DB
}

// NewGormDriverReviewRepository 创建生产环境司机评价仓储。
func NewGormDriverReviewRepository(db *gorm.DB) *GormDriverReviewRepository {
	return &GormDriverReviewRepository{db: db}
}

// ListPassengerReviewsByDriver 按司机 ID 分页查询其收到的乘客评价（order_review 表）。
func (r *GormDriverReviewRepository) ListPassengerReviewsByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]PassengerReview, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&PassengerReview{}).Where("driver_id = ?", driverID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []PassengerReview
	offset := int((page - 1) * pageSize)
	if offset < 0 {
		offset = 0
	}
	if err := r.db.WithContext(ctx).
		Select("id, order_id, user_id, driver_id, rating, comment, tags, UNIX_TIMESTAMP(created_at) as created_at").
		Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CreateDriverReview 写入司机对乘客的评价，依赖 order_id 唯一索引防止重复评价。
func (r *GormDriverReviewRepository) CreateDriverReview(ctx context.Context, review *DriverOrderReview) error {
	err := r.db.WithContext(ctx).Create(review).Error
	if err != nil && isDuplicateDriverReview(err) {
		return ErrReviewAlreadyExists
	}
	return err
}

// ListDriverReviewsByDriver 按司机 ID 分页查询其给出的乘客评价（driver_review 表）。
func (r *GormDriverReviewRepository) ListDriverReviewsByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]DriverOrderReview, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&DriverOrderReview{}).Where("driver_id = ?", driverID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []DriverOrderReview
	offset := int((page - 1) * pageSize)
	if offset < 0 {
		offset = 0
	}
	if err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// isDuplicateDriverReview 判断 MySQL 唯一键冲突。
func isDuplicateDriverReview(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062"))
}

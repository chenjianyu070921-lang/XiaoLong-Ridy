package svc

import (
	"context"

	"gorm.io/gorm"
)

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

// ReviewRepository 定义司机端查询收到的乘客评价的契约。
type ReviewRepository interface {
	ListPassengerReviewsByDriver(ctx context.Context, driverID int64, page, pageSize int32) ([]PassengerReview, int64, error)
	// CountAndAvgRatingByDriver 返回司机收到的评价总数与评分均值（服务平均分数据源）。
	CountAndAvgRatingByDriver(ctx context.Context, driverID int64) (count int64, avgRating float64, err error)
}

// GormDriverReviewRepository 是司机端评价的 MySQL/GORM 仓储实现。
type GormDriverReviewRepository struct {
	db *gorm.DB
}

// NewGormDriverReviewRepository 创建生产环境司机评价仓储。
func NewGormDriverReviewRepository(db *gorm.DB) *GormDriverReviewRepository {
	return &GormDriverReviewRepository{db: db}
}

// CountAndAvgRatingByDriver 统计司机收到的评价总数与评分均值（order_review 表）。
func (r *GormDriverReviewRepository) CountAndAvgRatingByDriver(ctx context.Context, driverID int64) (int64, float64, error) {
	var count int64
	var avgRating float64
	row := r.db.WithContext(ctx).
		Table("order_review").
		Select("COUNT(*) AS count, COALESCE(AVG(rating), 0) AS avg_rating").
		Where("driver_id = ?", driverID).
		Row()
	if err := row.Scan(&count, &avgRating); err != nil {
		return 0, 0, err
	}
	return count, avgRating, nil
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

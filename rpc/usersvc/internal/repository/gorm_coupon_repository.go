package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormCouponRepository 是基于 MySQL/GORM 的优惠券持久化仓储。
type gormCouponRepository struct {
	db *gorm.DB
}

// NewGormCouponRepository 创建生产环境使用的优惠券仓储。
func NewGormCouponRepository(db *gorm.DB) CouponRepository {
	return &gormCouponRepository{db: db}
}

// Claim 在事务内领取优惠券，确保库存和单用户领取上限一致。
func (r *gormCouponRepository) Claim(ctx context.Context, userID, couponID uint64) (*UserCouponWithTemplate, error) {
	var result *UserCouponWithTemplate
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var coupon model.Coupon
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", couponID).
			First(&coupon).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		if err != nil {
			return err
		}

		now := time.Now()
		if !isCouponAvailable(&coupon, now) {
			return ErrCouponUnavailable
		}
		if coupon.TotalCount > 0 && coupon.ReceivedCount >= coupon.TotalCount {
			return ErrCouponUnavailable
		}

		limit := coupon.PerUserLimit
		if limit <= 0 {
			limit = 1
		}
		var count int64
		if err := tx.Model(&model.UserCoupon{}).
			Where("user_id = ? AND coupon_id = ?", userID, couponID).
			Count(&count).Error; err != nil {
			return err
		}
		if count >= int64(limit) {
			return ErrCouponReceiveLimit
		}

		userCoupon := &model.UserCoupon{
			UserID:     userID,
			CouponID:   couponID,
			Status:     model.UserCouponStatusUnused,
			ReceivedAt: now,
			ExpireAt:   coupon.ValidEndAt,
		}
		if err := tx.Create(userCoupon).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Coupon{}).
			Where("id = ?", couponID).
			UpdateColumn("received_count", gorm.Expr("received_count + ?", 1)).Error; err != nil {
			return err
		}
		coupon.ReceivedCount++
		result = couponView(userCoupon, &coupon)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListByUser 查询用户优惠券列表，status 为 0 时返回全部状态。
// 排序字段 received_at, id 与迁移 12 的索引 idx_user_status_received 顺序一致，
// 使 WHERE user_id=? AND status=? 过滤与 ORDER BY 都走索引，避免 filesort。
func (r *gormCouponRepository) ListByUser(ctx context.Context, userID uint64, status int8) ([]*UserCouponWithTemplate, error) {
	var rows []couponListRow
	query := r.db.WithContext(ctx).
		Table("user_coupon AS uc").
		Select(`uc.id AS user_coupon_id, uc.user_id, uc.coupon_id, uc.order_id, uc.locked_order_id,
			uc.status, uc.received_at, uc.used_at, uc.locked_at, uc.expire_at, uc.created_at, uc.updated_at,
			c.name, c.type, c.face_value, c.discount, c.threshold_amount, c.car_type, c.city_code`).
		Joins("JOIN coupon AS c ON c.id = uc.coupon_id").
		Where("uc.user_id = ?", userID).
		Order("uc.received_at DESC, uc.id DESC")
	if status > 0 {
		query = query.Where("uc.status = ?", status)
	}
	// 列表查询无需返回全部历史，限制条数避免大结果集导致的回表与传输开销。
	query = query.Limit(200)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	list := make([]*UserCouponWithTemplate, 0, len(rows))
	for _, row := range rows {
		item := row.toView()
		if item.UserCoupon.Status == model.UserCouponStatusUnused && item.UserCoupon.ExpireAt.Before(now) {
			item.UserCoupon.Status = model.UserCouponStatusExpired
		}
		if status > 0 && item.UserCoupon.Status != status {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

// ListByUserPage 按用户和状态在数据库侧完成总数统计与分页查询，避免跨服务传输全量用户券。
func (r *gormCouponRepository) ListByUserPage(ctx context.Context, userID uint64, status int8, page, pageSize int) ([]*UserCouponWithTemplate, int64, error) {
	now := time.Now()
	offset := (page - 1) * pageSize

	countQuery := r.db.WithContext(ctx).
		Table("user_coupon AS uc").
		Where("uc.user_id = ?", userID)
	if status > 0 {
		countQuery = applyCouponStatusFilter(countQuery, status, now)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.WithContext(ctx).
		Table("user_coupon AS uc").
		Select(`uc.id AS user_coupon_id, uc.user_id, uc.coupon_id, uc.order_id, uc.locked_order_id,
			uc.status, uc.received_at, uc.used_at, uc.locked_at, uc.expire_at, uc.created_at, uc.updated_at,
			c.name, c.type, c.face_value, c.discount, c.threshold_amount, c.car_type, c.city_code`).
		Joins("JOIN coupon AS c ON c.id = uc.coupon_id").
		Where("uc.user_id = ?", userID).
		Order("uc.received_at DESC, uc.id DESC").
		Offset(offset).
		Limit(pageSize)
	if status > 0 {
		query = applyCouponStatusFilter(query, status, now)
	}

	var rows []couponListRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	list := make([]*UserCouponWithTemplate, 0, len(rows))
	for _, row := range rows {
		item := row.toView()
		if item.UserCoupon.Status == model.UserCouponStatusUnused && item.UserCoupon.ExpireAt.Before(now) {
			item.UserCoupon.Status = model.UserCouponStatusExpired
		}
		list = append(list, item)
	}
	return list, total, nil
}

// applyCouponStatusFilter 将用户券状态筛选转换为数据库条件。
// 未使用券过期后在业务层展示为过期，因此状态 3 需要同时匹配数据库中的过期状态
// 和已超过有效期的未使用券，保证 COUNT 与分页列表的口径一致。
func applyCouponStatusFilter(query *gorm.DB, status int8, now time.Time) *gorm.DB {
	switch status {
	case model.UserCouponStatusUnused:
		return query.Where("uc.status = ? AND uc.expire_at >= ?", status, now)
	case model.UserCouponStatusExpired:
		return query.Where("(uc.status = ? OR (uc.status = ? AND uc.expire_at < ?))", status, model.UserCouponStatusUnused, now)
	default:
		return query.Where("uc.status = ?", status)
	}
}

// Lock 将用户券从未使用状态锁定到指定订单。
func (r *gormCouponRepository) Lock(ctx context.Context, userID, userCouponID, orderID uint64, carType int8, cityCode string) (*UserCouponWithTemplate, error) {
	var result *UserCouponWithTemplate
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCoupon model.UserCoupon
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", userCouponID, userID).
			First(&userCoupon).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserCouponNotFound
		}
		if err != nil {
			return err
		}

		var coupon model.Coupon
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userCoupon.CouponID).
			First(&coupon).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		if err != nil {
			return err
		}

		now := time.Now()
		if !isCouponUsable(&userCoupon, &coupon, now, carType, cityCode) {
			return ErrCouponUnavailable
		}
		userCoupon.Status = model.UserCouponStatusLocked
		userCoupon.LockedOrderID = orderID
		userCoupon.LockedAt = &now
		userCoupon.UpdatedAt = now
		if err := tx.Save(&userCoupon).Error; err != nil {
			return err
		}
		result = couponView(&userCoupon, &coupon)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Release 将指定订单锁定的用户券释放回未使用状态。
func (r *gormCouponRepository) Release(ctx context.Context, userID, userCouponID, orderID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCoupon model.UserCoupon
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", userCouponID, userID).
			First(&userCoupon).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserCouponNotFound
		}
		if err != nil {
			return err
		}
		if userCoupon.Status != model.UserCouponStatusLocked || userCoupon.LockedOrderID != orderID {
			return ErrCouponUnavailable
		}
		return tx.Model(&model.UserCoupon{}).
			Where("id = ? AND user_id = ?", userCouponID, userID).
			Updates(map[string]interface{}{
				"status":          model.UserCouponStatusUnused,
				"locked_order_id": 0,
				"locked_at":       nil,
				"updated_at":      time.Now(),
			}).Error
	})
}

// ReleaseByOrder 将指定订单锁定的用户券恢复为未使用状态，供取消订单幂等回滚使用。
func (r *gormCouponRepository) ReleaseByOrder(ctx context.Context, userID, orderID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(&model.UserCoupon{}).
			Where("user_id = ? AND locked_order_id = ? AND status = ?", userID, orderID, model.UserCouponStatusLocked).
			Updates(map[string]interface{}{
				"status":          model.UserCouponStatusUnused,
				"locked_order_id": 0,
				"locked_at":       nil,
				"updated_at":      time.Now(),
			}).Error
	})
}

// couponListRow 承接 user_coupon 与 coupon 关联查询结果。
// ConsumeByOrder 将指定订单锁定的用户券核销为已使用状态。
// 该方法用于支付成功后的最终确认，重复消费同一订单时保持幂等。
func (r *gormCouponRepository) ConsumeByOrder(ctx context.Context, userID, orderID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCoupon model.UserCoupon
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND (locked_order_id = ? OR order_id = ?)", userID, orderID, orderID).
			First(&userCoupon).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if userCoupon.Status == model.UserCouponStatusUsed && userCoupon.OrderID == orderID {
			return nil
		}
		if userCoupon.Status != model.UserCouponStatusLocked || userCoupon.LockedOrderID != orderID {
			return ErrCouponUnavailable
		}
		now := time.Now()
		return tx.Model(&model.UserCoupon{}).
			Where("id = ? AND user_id = ?", userCoupon.ID, userID).
			Updates(map[string]interface{}{
				"status":          model.UserCouponStatusUsed,
				"order_id":        orderID,
				"locked_order_id": 0,
				"locked_at":       nil,
				"used_at":         now,
				"updated_at":      now,
			}).Error
	})
}

type couponListRow struct {
	UserCouponID    uint64
	UserID          uint64
	CouponID        uint64
	OrderID         uint64
	LockedOrderID   uint64
	Status          int8
	ReceivedAt      time.Time
	UsedAt          *time.Time
	LockedAt        *time.Time
	ExpireAt        time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Type            int8
	FaceValue       float64
	Discount        float64
	ThresholdAmount float64
	CarType         int8
	CityCode        string
}

// isCouponUsable 判断用户券是否可被本次下单锁定。
func isCouponUsable(userCoupon *model.UserCoupon, coupon *model.Coupon, now time.Time, carType int8, cityCode string) bool {
	if userCoupon == nil || coupon == nil || userCoupon.Status != model.UserCouponStatusUnused {
		return false
	}
	if userCoupon.ExpireAt.Before(now) || !isCouponAvailable(coupon, now) {
		return false
	}
	if coupon.CarType != model.CouponCarTypeAll && coupon.CarType != carType {
		return false
	}
	cityCode = strings.TrimSpace(cityCode)
	return coupon.CityCode == "" || cityCode == "" || coupon.CityCode == cityCode
}

// toView 将关联查询行转换为统一视图。
func (r couponListRow) toView() *UserCouponWithTemplate {
	return &UserCouponWithTemplate{
		UserCoupon: &model.UserCoupon{
			ID:            r.UserCouponID,
			UserID:        r.UserID,
			CouponID:      r.CouponID,
			OrderID:       r.OrderID,
			LockedOrderID: r.LockedOrderID,
			Status:        r.Status,
			ReceivedAt:    r.ReceivedAt,
			UsedAt:        r.UsedAt,
			LockedAt:      r.LockedAt,
			ExpireAt:      r.ExpireAt,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		},
		Coupon: &model.Coupon{
			ID:              r.CouponID,
			Name:            r.Name,
			Type:            r.Type,
			FaceValue:       r.FaceValue,
			Discount:        r.Discount,
			ThresholdAmount: r.ThresholdAmount,
			CarType:         r.CarType,
			CityCode:        r.CityCode,
		},
	}
}

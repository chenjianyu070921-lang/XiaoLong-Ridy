package repository

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DriverPunishmentRepository 定义司机处罚相关的数据访问接口，
// 覆盖 driver、driver_location、driver_score、driver_punishment_effect 四张表。
// 所有方法都接受 tx *gorm.DB，便于调用方把多个处罚动作包在同一个事务里保证原子性。
type DriverPunishmentRepository interface {
	// EffectExists 判断某个处罚动作的生效记录是否已存在，用于幂等跳过。
	EffectExists(ctx context.Context, tx *gorm.DB, eventID, actionType string) (bool, error)
	// UpdateStatus 更新 driver 表的账号状态（冻结/恢复）。
	UpdateStatus(ctx context.Context, tx *gorm.DB, driverID uint64, status int8, now time.Time) error
	// UpdateOnlineStatus 更新 driver 表的在线状态（停止派单/恢复派单）。
	UpdateOnlineStatus(ctx context.Context, tx *gorm.DB, driverID uint64, onlineStatus int8, now time.Time) error
	// UpdateLocationOnlineStatus 同步 driver_location 表的在线状态。
	UpdateLocationOnlineStatus(ctx context.Context, tx *gorm.DB, driverID uint64, onlineStatus int8) error
	// AddScore 增减服务分，扣分后不低于 0。
	AddScore(ctx context.Context, tx *gorm.DB, driverID uint64, delta float64, now time.Time) error
	// AddLevel 增减司机等级，降级后不低于 1。
	AddLevel(ctx context.Context, tx *gorm.DB, driverID uint64, delta int8, now time.Time) error
	// RecordEffect 写入处罚生效记录；已存在时静默跳过。
	RecordEffect(ctx context.Context, tx *gorm.DB, effect *model.DriverPunishmentEffect) error
}

// gormPunishmentRepository 是基于 GORM 的处罚数据访问实现。
type gormPunishmentRepository struct {
	db *gorm.DB
}

// NewGormPunishmentRepository 创建处罚数据访问实例。
func NewGormPunishmentRepository(db *gorm.DB) DriverPunishmentRepository {
	return &gormPunishmentRepository{db: db}
}

// withTx 返回实际执行的数据句柄：优先使用调用方传入的事务，未传时回落到实例连接。
func (r *gormPunishmentRepository) withTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// EffectExists 判断某个处罚动作的生效记录是否已存在，用于幂等跳过。
func (r *gormPunishmentRepository) EffectExists(ctx context.Context, tx *gorm.DB, eventID, actionType string) (bool, error) {
	var count int64
	err := r.withTx(tx).WithContext(ctx).
		Model(&model.DriverPunishmentEffect{}).
		Where("event_id = ? AND action_type = ?", eventID, actionType).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateStatus 更新 driver 表的账号状态（冻结/恢复）。
func (r *gormPunishmentRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, driverID uint64, status int8, now time.Time) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.Driver{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}).Error
}

// UpdateOnlineStatus 更新 driver 表的在线状态（停止派单/恢复派单）。
func (r *gormPunishmentRepository) UpdateOnlineStatus(ctx context.Context, tx *gorm.DB, driverID uint64, onlineStatus int8, now time.Time) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.Driver{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"online_status": onlineStatus,
			"updated_at":    now,
		}).Error
}

// UpdateLocationOnlineStatus 同步 driver_location 表的在线状态。
func (r *gormPunishmentRepository) UpdateLocationOnlineStatus(ctx context.Context, tx *gorm.DB, driverID uint64, onlineStatus int8) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.DriverLocation{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"online_status": onlineStatus,
		}).Error
}

// AddScore 增减服务分，扣分后不低于 0。
func (r *gormPunishmentRepository) AddScore(ctx context.Context, tx *gorm.DB, driverID uint64, delta float64, now time.Time) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.DriverScore{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"score":      gorm.Expr("GREATEST(0, score + ?)", delta),
			"updated_at": now,
		}).Error
}

// AddLevel 增减司机等级，降级后不低于 1。
func (r *gormPunishmentRepository) AddLevel(ctx context.Context, tx *gorm.DB, driverID uint64, delta int8, now time.Time) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.DriverScore{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"level":      gorm.Expr("GREATEST(1, level + ?)", delta),
			"updated_at": now,
		}).Error
}

// RecordEffect 写入处罚生效记录；已存在时静默跳过。
func (r *gormPunishmentRepository) RecordEffect(ctx context.Context, tx *gorm.DB, effect *model.DriverPunishmentEffect) error {
	return r.withTx(tx).WithContext(ctx).
		Model(&model.DriverPunishmentEffect{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(effect).Error
}

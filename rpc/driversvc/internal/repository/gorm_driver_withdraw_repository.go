package repository

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"

	"gorm.io/gorm"
)

type gormDriverWithdrawRepository struct {
	db *gorm.DB
}

// NewGormDriverWithdrawRepository 创建基于 GORM 的提现仓储。
func NewGormDriverWithdrawRepository(db *gorm.DB) DriverWithdrawRepository {
	return &gormDriverWithdrawRepository{db: db}
}

// Create 写入一条提现申请记录。
func (r *gormDriverWithdrawRepository) Create(ctx context.Context, withdraw *model.DriverWithdraw) error {
	return r.db.WithContext(ctx).Create(withdraw).Error
}

// ListByDriver 按司机 ID 分页查询提现记录，按申请时间倒序返回本页与总数。
func (r *gormDriverWithdrawRepository) ListByDriver(ctx context.Context, driverID uint64, page, pageSize int32) ([]*model.DriverWithdraw, int64, error) {
	page = clampPage(page)
	pageSize = clampPageSize(pageSize)

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.DriverWithdraw{}).
		Where("driver_id = ?", driverID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []*model.DriverWithdraw
	if err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("applied_at DESC, id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// adminWithdrawScope 将管理后台筛选条件转换为查询作用域，供计数与取数复用。
func adminWithdrawScope(filter AdminWithdrawFilter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filter.Status > 0 {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.DriverID > 0 {
			db = db.Where("driver_id = ?", filter.DriverID)
		}
		if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
			kw := "%" + keyword + "%"
			db = db.Where("withdraw_no LIKE ? OR payee_name LIKE ? OR pay_account LIKE ?", kw, kw, kw)
		}
		return db
	}
}

// AdminList 按管理后台筛选条件分页查询提现记录，按申请时间倒序返回本页与总数。
func (r *gormDriverWithdrawRepository) AdminList(ctx context.Context, filter AdminWithdrawFilter) ([]*model.DriverWithdraw, int64, error) {
	page := clampPage(filter.Page)
	pageSize := clampPageSize(filter.PageSize)

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.DriverWithdraw{}).
		Scopes(adminWithdrawScope(filter)).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []*model.DriverWithdraw
	if err := r.db.WithContext(ctx).
		Scopes(adminWithdrawScope(filter)).
		Order("applied_at DESC, id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// GetByID 按主键查询提现记录。
func (r *gormDriverWithdrawRepository) GetByID(ctx context.Context, id uint64) (*model.DriverWithdraw, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var record model.DriverWithdraw
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// Audit 写入提现审核结果：status 为目标状态，remark 为审核备注，paidAt 仅打款成功时写入。
func (r *gormDriverWithdrawRepository) Audit(ctx context.Context, id uint64, status int32, remark string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
		"remark": remark,
	}
	if paidAt != nil {
		updates["paid_at"] = *paidAt
	}
	return r.db.WithContext(ctx).
		Model(&model.DriverWithdraw{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// clampPage 页码收敛：至少 1。
func clampPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

// clampPageSize 每页条数收敛：默认 20，上限 100。
func clampPageSize(pageSize int32) int32 {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

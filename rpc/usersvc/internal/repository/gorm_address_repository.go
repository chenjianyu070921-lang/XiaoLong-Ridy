package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"

	"gorm.io/gorm"
)

// gormAddressRepository 是基于 MySQL/GORM 的乘客常用地址仓储实现。
type gormAddressRepository struct {
	db *gorm.DB
}

// NewGormAddressRepository 创建生产环境使用的常用地址持久化仓储。
func NewGormAddressRepository(db *gorm.DB) AddressRepository {
	return &gormAddressRepository{db: db}
}

// Create 保存新的常用地址，并维护同一用户只有一个默认地址。
func (r *gormAddressRepository) Create(ctx context.Context, address *model.UserAddress) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if address.IsDefault == model.UserAddressIsDefault {
			if err := clearDefaultAddress(ctx, tx, address.UserID, 0); err != nil {
				return err
			}
		}
		return tx.Create(address).Error
	})
}

// ListByUser 按默认地址、排序值和 ID 顺序返回乘客未删除的常用地址。
func (r *gormAddressRepository) ListByUser(ctx context.Context, userID uint64) ([]*model.UserAddress, error) {
	var list []*model.UserAddress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, sort DESC, id ASC").
		Find(&list).Error
	return list, err
}

// FindByID 查询指定用户自己的常用地址。
func (r *gormAddressRepository) FindByID(ctx context.Context, userID, addressID uint64) (*model.UserAddress, error) {
	var address model.UserAddress
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAddressNotFound
	}
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// Update 更新乘客自己的常用地址，并在设置默认地址时清理旧默认地址。
func (r *gormAddressRepository) Update(ctx context.Context, address *model.UserAddress) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if address.IsDefault == model.UserAddressIsDefault {
			if err := clearDefaultAddress(ctx, tx, address.UserID, address.ID); err != nil {
				return err
			}
		}
		res := tx.Model(&model.UserAddress{}).
			Where("id = ? AND user_id = ?", address.ID, address.UserID).
			Updates(map[string]interface{}{
				"contact_name":  address.ContactName,
				"contact_phone": address.ContactPhone,
				"tag":           address.Tag,
				"address":       address.Address,
				"longitude":     address.Longitude,
				"latitude":      address.Latitude,
				"is_default":    address.IsDefault,
				"sort":          address.Sort,
				"updated_at":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrAddressNotFound
		}
		return tx.Where("id = ? AND user_id = ?", address.ID, address.UserID).First(address).Error
	})
}

// Delete 软删除乘客自己的常用地址。
func (r *gormAddressRepository) Delete(ctx context.Context, userID, addressID uint64) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", addressID, userID).
		Updates(&model.UserAddress{UpdatedAt: now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAddressNotFound
	}
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", addressID, userID).Delete(&model.UserAddress{}).Error
}

// clearDefaultAddress 清除同一用户下除 keepID 外的默认地址标记。
func clearDefaultAddress(ctx context.Context, tx *gorm.DB, userID, keepID uint64) error {
	query := tx.WithContext(ctx).Model(&model.UserAddress{}).
		Where("user_id = ?", userID)
	if keepID > 0 {
		query = query.Where("id <> ?", keepID)
	}
	return query.Updates(map[string]interface{}{
		"is_default": model.UserAddressNotDefault,
		"updated_at": time.Now(),
	}).Error
}

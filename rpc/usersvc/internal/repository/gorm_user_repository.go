package repository

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// gormUserRepository 是基于 MySQL/GORM 的乘客用户仓储实现。
type gormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository 创建生产环境使用的用户持久化仓储。
func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

// FindByPhone 根据手机号查询未软删除的用户。
func (r *gormUserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ? AND deleted_at IS NULL", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据用户 ID 查询未软删除的用户资料。
func (r *gormUserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 将首次短信登录的乘客账号写入 user 表。
func (r *gormUserRepository) Create(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if isDuplicateUserKey(err) {
		return ErrPhoneExists
	}
	return err
}

// Update 保存用户资料变更，并保持手机号唯一约束错误语义一致。
func (r *gormUserRepository) Update(ctx context.Context, user *model.User) error {
	updates := map[string]interface{}{
		"phone":           user.Phone,
		"password_hash":   user.PasswordHash,
		"nickname":        user.Nickname,
		"avatar_url":      user.AvatarURL,
		"gender":          user.Gender,
		"real_name":       user.RealName,
		"id_card_no":      user.IDCardNo,
		"register_source": user.RegisterSource,
		"status":          user.Status,
		"updated_at":      time.Now(),
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.User{}).
			Where("id = ? AND deleted_at IS NULL", user.ID).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrUserNotFound
		}
		return tx.Where("id = ? AND deleted_at IS NULL", user.ID).First(user).Error
	})
	if isDuplicateUserKey(err) {
		return ErrPhoneExists
	}
	return err
}

// isDuplicateUserKey 判断 MySQL 返回值是否为唯一键冲突。
func isDuplicateUserKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

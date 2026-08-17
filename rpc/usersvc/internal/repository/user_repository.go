package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrPhoneExists  = errors.New("phone already registered")
)

type UserRepository interface {
	// FindByPhone 根据手机号查询用户。
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	// FindByID 根据用户 ID 查询用户资料。
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	// Create 持久化新增用户。
	Create(ctx context.Context, user *model.User) error
	// Update 保存用户资料变更，当前用于个人中心和实名信息更新。
	Update(ctx context.Context, user *model.User) error
}

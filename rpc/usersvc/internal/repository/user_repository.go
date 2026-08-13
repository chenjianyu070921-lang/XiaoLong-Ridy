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
	// Create 持久化新增用户。
	Create(ctx context.Context, user *model.User) error
}

package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

var (
	// ErrAddressNotFound 表示用户常用地址不存在或不属于当前用户。
	ErrAddressNotFound = errors.New("address not found")
)

// AddressRepository 定义用户常用地址仓储契约。
type AddressRepository interface {
	Create(ctx context.Context, address *model.UserAddress) error
	ListByUser(ctx context.Context, userID uint64) ([]*model.UserAddress, error)
	FindByID(ctx context.Context, userID, addressID uint64) (*model.UserAddress, error)
	Update(ctx context.Context, address *model.UserAddress) error
	Delete(ctx context.Context, userID, addressID uint64) error
}

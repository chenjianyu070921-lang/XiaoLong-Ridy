package repository

import (
	"context"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

// MemoryAddressRepository 是本地开发和测试使用的常用地址内存仓储。
type MemoryAddressRepository struct {
	mu        sync.RWMutex
	nextID    uint64
	addresses map[uint64]*model.UserAddress
}

// NewMemoryAddressRepository 创建常用地址内存仓储。
func NewMemoryAddressRepository() *MemoryAddressRepository {
	return &MemoryAddressRepository{
		nextID:    1,
		addresses: make(map[uint64]*model.UserAddress),
	}
}

// Create 保存新的常用地址，并在需要时维护同一用户下只有一个默认地址。
func (r *MemoryAddressRepository) Create(_ context.Context, address *model.UserAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	copied := *address
	copied.ID = r.nextID
	copied.CreatedAt = now
	copied.UpdatedAt = now
	r.nextID++

	if copied.IsDefault == model.UserAddressIsDefault {
		r.clearDefaultLocked(copied.UserID, copied.ID)
	}
	r.addresses[copied.ID] = &copied

	address.ID = copied.ID
	address.CreatedAt = copied.CreatedAt
	address.UpdatedAt = copied.UpdatedAt
	return nil
}

// ListByUser 返回指定用户未删除的常用地址列表。
func (r *MemoryAddressRepository) ListByUser(_ context.Context, userID uint64) ([]*model.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*model.UserAddress, 0)
	for _, address := range r.addresses {
		if address.UserID != userID || address.DeletedAt.Valid {
			continue
		}
		copied := *address
		list = append(list, &copied)
	}
	return list, nil
}

// FindByID 查询当前用户自己的常用地址。
func (r *MemoryAddressRepository) FindByID(_ context.Context, userID, addressID uint64) (*model.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	address, ok := r.addresses[addressID]
	if !ok || address.UserID != userID || address.DeletedAt.Valid {
		return nil, ErrAddressNotFound
	}
	copied := *address
	return &copied, nil
}

// Update 更新当前用户自己的常用地址。
func (r *MemoryAddressRepository) Update(_ context.Context, address *model.UserAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.addresses[address.ID]
	if !ok || current.UserID != address.UserID || current.DeletedAt.Valid {
		return ErrAddressNotFound
	}
	if address.IsDefault == model.UserAddressIsDefault {
		r.clearDefaultLocked(address.UserID, address.ID)
	}

	current.ContactName = address.ContactName
	current.ContactPhone = address.ContactPhone
	current.Tag = address.Tag
	current.Address = address.Address
	current.Longitude = address.Longitude
	current.Latitude = address.Latitude
	current.IsDefault = address.IsDefault
	current.UpdatedAt = time.Now()

	copied := *current
	*address = copied
	return nil
}

// Delete 软删除当前用户自己的常用地址。
func (r *MemoryAddressRepository) Delete(_ context.Context, userID, addressID uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	address, ok := r.addresses[addressID]
	if !ok || address.UserID != userID || address.DeletedAt.Valid {
		return ErrAddressNotFound
	}
	now := time.Now()
	address.DeletedAt.Time = now
	address.DeletedAt.Valid = true
	address.UpdatedAt = now
	return nil
}

func (r *MemoryAddressRepository) clearDefaultLocked(userID, keepID uint64) {
	for _, address := range r.addresses {
		if address.UserID == userID && address.ID != keepID && !address.DeletedAt.Valid {
			address.IsDefault = model.UserAddressNotDefault
		}
	}
}

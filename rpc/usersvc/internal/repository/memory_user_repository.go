package repository

import (
	"context"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
)

type MemoryUserRepository struct {
	mu           sync.RWMutex
	nextID       uint64
	usersByID    map[uint64]*model.User
	usersByPhone map[string]*model.User
}

// NewMemoryUserRepository 创建用于本地开发和测试的内存仓储。
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		nextID:       1,
		usersByID:    make(map[uint64]*model.User),
		usersByPhone: make(map[string]*model.User),
	}
}

// FindByPhone 返回用户副本，避免调用方直接修改仓储内部状态。
func (r *MemoryUserRepository) FindByPhone(_ context.Context, phone string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByPhone[phone]
	if !ok {
		return nil, ErrUserNotFound
	}

	copied := *user
	return &copied, nil
}

// Create 插入新用户并分配自增 ID。
func (r *MemoryUserRepository) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.usersByPhone[user.Phone]; ok {
		return ErrPhoneExists
	}

	now := time.Now()
	copied := *user
	copied.ID = r.nextID
	copied.CreatedAt = now
	copied.UpdatedAt = now

	r.nextID++
	r.usersByID[copied.ID] = &copied
	r.usersByPhone[copied.Phone] = &copied

	user.ID = copied.ID
	user.CreatedAt = copied.CreatedAt
	user.UpdatedAt = copied.UpdatedAt
	return nil
}

package repository

import (
	"context"
	"sort"
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

// ListPage 按 ID 倒序返回内存用户分页结果，模拟生产仓储的查询语义。
func (r *MemoryUserRepository) ListPage(_ context.Context, status, page, pageSize int) ([]*model.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]*model.User, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		if status > 0 && user.Status != status {
			continue
		}
		copied := *user
		items = append(items, &copied)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []*model.User{}, int64(len(items)), nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], int64(len(items)), nil
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

// FindByID 返回指定用户 ID 的资料副本，供个人中心和实名服务查询。
func (r *MemoryUserRepository) FindByID(_ context.Context, id uint64) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
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

// Update 保存用户资料副本，确保手机号索引和 ID 索引保持一致。
func (r *MemoryUserRepository) Update(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.usersByID[user.ID]
	if !ok {
		return ErrUserNotFound
	}
	if current.Phone != user.Phone {
		if _, exists := r.usersByPhone[user.Phone]; exists {
			return ErrPhoneExists
		}
		delete(r.usersByPhone, current.Phone)
	}

	copied := *user
	copied.CreatedAt = current.CreatedAt
	copied.UpdatedAt = time.Now()
	r.usersByID[copied.ID] = &copied
	r.usersByPhone[copied.Phone] = &copied

	user.CreatedAt = copied.CreatedAt
	user.UpdatedAt = copied.UpdatedAt
	return nil
}

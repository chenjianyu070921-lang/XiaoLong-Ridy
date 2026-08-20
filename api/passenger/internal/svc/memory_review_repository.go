package svc

import (
	"context"
	"sync"
	"time"
)

type MemoryReviewRepository struct {
	mu      sync.Mutex
	nextID  uint64
	byOrder map[uint64]*OrderReview
}

func NewMemoryReviewRepository() *MemoryReviewRepository {
	return &MemoryReviewRepository{nextID: 1, byOrder: make(map[uint64]*OrderReview)}
}

func (r *MemoryReviewRepository) Create(_ context.Context, review *OrderReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byOrder[review.OrderID]; ok {
		return ErrReviewAlreadyExists
	}
	now := time.Now()
	copied := *review
	copied.ID = r.nextID
	if copied.CreatedAt.IsZero() {
		copied.CreatedAt = now
	}
	copied.UpdatedAt = now
	r.nextID++
	r.byOrder[copied.OrderID] = &copied
	*review = copied
	return nil
}

package aiagent

import (
	"context"
	"sync"
)

// Conversation 表示一个管理员在某场景下的会话上下文。
// 会话绑定单一场景，只保留脱敏后的轮次摘要。
type Conversation struct {
	ID        string
	AdminID   int64
	Scene     Scene
	Rounds    []Round
	UpdatedAt int64
}

// Round 表示一轮问答的脱敏摘要。
type Round struct {
	Question string
	Answer   Answer
}

// ConversationSummary 是会话列表项的脱敏摘要。
type ConversationSummary struct {
	ConversationID string
	Scene          Scene
	SourceMode     SourceMode
	Summary        string
	UpdatedAt      int64
}

// SessionStore 是会话热上下文的存取端口。
// 生产实现将热上下文存 Redis 并带 TTL，历史/审计摘要落库。
type SessionStore interface {
	Load(ctx context.Context, conversationID string) (*Conversation, error)
	Save(ctx context.Context, c *Conversation) error
	Delete(ctx context.Context, conversationID string) error
	// ListByAdmin 返回某管理员拥有的全部会话。
	ListByAdmin(ctx context.Context, adminID int64) ([]*Conversation, error)
}

// InMemoryStore 是进程内会话存储，供单元测试与单机演示使用。
type InMemoryStore struct {
	mu sync.Mutex
	m  map[string]*Conversation
}

// NewInMemoryStore 构造内存会话存储。
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{m: make(map[string]*Conversation)}
}

// Load 实现 SessionStore。
func (s *InMemoryStore) Load(_ context.Context, id string) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[id], nil
}

// Save 实现 SessionStore。
func (s *InMemoryStore) Save(_ context.Context, c *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[c.ID] = c
	return nil
}

// Delete 实现 SessionStore。
func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

// ListByAdmin 实现 SessionStore。
func (s *InMemoryStore) ListByAdmin(_ context.Context, adminID int64) ([]*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Conversation, 0)
	for _, c := range s.m {
		if c.AdminID == adminID {
			out = append(out, c)
		}
	}
	return out, nil
}

// appendRound 追加一轮并截断到 maxRounds，超过按 FIFO 丢弃最早轮次。
func appendRound(rounds []Round, r Round, maxRounds int) []Round {
	if maxRounds <= 0 {
		maxRounds = 6
	}
	rounds = append(rounds, r)
	if len(rounds) > maxRounds {
		rounds = rounds[len(rounds)-maxRounds:]
	}
	return rounds
}

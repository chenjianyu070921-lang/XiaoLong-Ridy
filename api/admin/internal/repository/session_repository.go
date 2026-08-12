package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"XiaoLong-Ridy/api/admin/internal/model"
	"github.com/redis/go-redis/v9"
)

// SessionRepository 封装管理员会话在 Redis 中的保存、读取和删除操作。
type SessionRepository struct {
	redis       *redis.Client
	tokenPrefix string
	sessionTTL  time.Duration
}

// NewSessionRepository 使用指定的键前缀和有效期创建会话仓储。
func NewSessionRepository(redisClient *redis.Client, tokenPrefix string, sessionTTL time.Duration) *SessionRepository {
	return &SessionRepository{
		redis:       redisClient,
		tokenPrefix: tokenPrefix,
		sessionTTL:  sessionTTL,
	}
}

// Save 将管理员会话序列化后写入 Redis，并设置统一过期时间。
func (r *SessionRepository) Save(ctx context.Context, session model.AdminSession) error {
	b, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	return r.redis.Set(ctx, r.key(session.Token), b, r.sessionTTL).Err()
}

// Get 根据访问令牌从 Redis 读取并反序列化管理员会话。
func (r *SessionRepository) Get(ctx context.Context, token string) (*model.AdminSession, error) {
	val, err := r.redis.Get(ctx, r.key(token)).Bytes()
	if err != nil {
		return nil, err
	}
	var session model.AdminSession
	if err := json.Unmarshal(val, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

// Delete 删除已退出登录或需要失效的管理员会话。
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	return r.redis.Del(ctx, r.key(token)).Err()
}

// key 统一生成 Redis 会话键，避免与其他业务模块发生键冲突。
func (r *SessionRepository) key(token string) string {
	return r.tokenPrefix + token
}

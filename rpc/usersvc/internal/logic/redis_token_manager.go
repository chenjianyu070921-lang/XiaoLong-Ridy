package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/jwtx"

	"github.com/redis/go-redis/v9"
)

// RedisTokenManager 使用 Redis 保存 Refresh Token 会话和注销黑名单。
type RedisTokenManager struct {
	client     *redis.Client
	signingKey string
}

// redisCallContext 为 Redis 操作设置统一短超时，避免 Redis 故障拖垮 RPC 请求。
func redisCallContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// redisRefreshSession 是 Redis 中保存的刷新令牌会话结构。
type redisRefreshSession struct {
	UserID     uint64    `json:"userId"`
	Phone      string    `json:"phone"`
	UserStatus int       `json:"userStatus"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// NewRedisTokenManager 创建生产环境使用的令牌管理器。
func NewRedisTokenManager(client *redis.Client, signingKey string) *RedisTokenManager {
	return &RedisTokenManager{
		client:     client,
		signingKey: signingKey,
	}
}

// Issue 为用户签发 Access Token，并把 Refresh Token 会话写入 Redis。
func (m *RedisTokenManager) Issue(userID uint64, phone string, userStatus int) (string, string, error) {
	accessToken, err := jwtx.SignUserToken(jwtx.UserTokenPayload{
		UserID:     userID,
		Phone:      phone,
		Role:       "passenger",
		UserStatus: userStatus,
		Issuer:     "huaxiaozhu-api",
		TTL:        accessTokenTTL,
	}, m.signingKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := randomToken()
	if err != nil {
		return "", "", err
	}
	session := redisRefreshSession{
		UserID:     userID,
		Phone:      phone,
		UserStatus: userStatus,
		ExpiresAt:  time.Now().Add(refreshTokenTTL),
	}
	payload, err := json.Marshal(session)
	if err != nil {
		return "", "", err
	}
	ctx, cancel := redisCallContext()
	defer cancel()
	if err := m.client.Set(ctx, refreshTokenKey(refreshToken), payload, refreshTokenTTL).Err(); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// Refresh 先读取旧 Refresh Token，确认新令牌签发成功后再删除旧令牌，
// 避免下游 Redis/签发失败时误删用户仍可用的刷新令牌。
func (m *RedisTokenManager) Refresh(refreshToken string) (string, string, error) {
	ctx, cancel := redisCallContext()
	defer cancel()
	payload, err := m.client.Get(ctx, refreshTokenKey(refreshToken)).Bytes()
	if errors.Is(err, redis.Nil) {
		return "", "", ErrInvalidToken
	}
	if err != nil {
		return "", "", err
	}

	var session redisRefreshSession
	if err := json.Unmarshal(payload, &session); err != nil {
		return "", "", err
	}
	if time.Now().After(session.ExpiresAt) {
		return "", "", ErrTokenExpired
	}
	token, nextRefresh, err := m.Issue(session.UserID, session.Phone, session.UserStatus)
	if err != nil {
		return "", "", err
	}
	if err := m.client.Del(ctx, refreshTokenKey(refreshToken)).Err(); err != nil {
		return "", "", err
	}
	return token, nextRefresh, nil
}

// Revoke 将当前 Access Token 写入 Redis 黑名单直到其自然过期。
func (m *RedisTokenManager) Revoke(token string) error {
	if _, err := m.Validate(token); err != nil {
		return err
	}
	ctx, cancel := redisCallContext()
	defer cancel()
	return m.client.Set(ctx, revokedTokenKey(token), "1", accessTokenTTL).Err()
}

// Validate 校验 Access Token 签名、有效期和 Redis 注销黑名单。
func (m *RedisTokenManager) Validate(token string) (*jwtx.UserClaims, error) {
	claims, err := jwtx.ParseUserToken(token, m.signingKey)
	if errors.Is(err, jwtx.ErrInvalidToken) {
		return nil, ErrInvalidToken
	}
	if errors.Is(err, jwtx.ErrTokenExpired) {
		return nil, ErrTokenExpired
	}
	if err != nil {
		return nil, err
	}

	ctx, cancel := redisCallContext()
	defer cancel()
	exists, err := m.client.Exists(ctx, revokedTokenKey(token)).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// refreshTokenKey 返回 Refresh Token 会话 Redis key。
func refreshTokenKey(token string) string {
	return fmt.Sprintf("usersvc:token:refresh:%s", token)
}

// revokedTokenKey 返回 Access Token 注销黑名单 Redis key。
func revokedTokenKey(token string) string {
	return fmt.Sprintf("usersvc:token:revoked:%s", token)
}

package svc

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CodeCache 验证码存储抽象：本地联调用本地内存，多实例/生产用 Redis 共享。
type CodeCache interface {
	TTL() time.Duration
	Set(phone, code string)
	Verify(phone, code string) bool
}

// ---------- 本地内存实现（单实例联调） ----------

type codeEntry struct {
	code      string
	expiresAt time.Time
}

// LocalCodeCache 本地内存验证码存储。仅适用于单实例本地联调。
type LocalCodeCache struct {
	mu      sync.RWMutex
	entries map[string]codeEntry
	ttl     time.Duration
}

func NewLocalCodeCache(ttl time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		entries: make(map[string]codeEntry),
		ttl:     ttl,
	}
}

func (c *LocalCodeCache) TTL() time.Duration { return c.ttl }

func (c *LocalCodeCache) Set(phone, code string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[phone] = codeEntry{code: code, expiresAt: time.Now().Add(c.ttl)}
}

func (c *LocalCodeCache) Verify(phone, code string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[phone]
	if !ok {
		return false
	}
	delete(c.entries, phone)
	return entry.code == code && time.Now().Before(entry.expiresAt)
}

// ---------- Redis 实现（多实例/生产） ----------

// RedisCodeCache 基于 Redis 的验证码存储，支持多实例共享、自动过期。
type RedisCodeCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedisCodeCache(rdb *redis.Client, ttl time.Duration) *RedisCodeCache {
	return &RedisCodeCache{rdb: rdb, ttl: ttl}
}

func (c *RedisCodeCache) TTL() time.Duration { return c.ttl }

func (c *RedisCodeCache) key(phone string) string {
	return "driver:sms:code:" + phone
}

func (c *RedisCodeCache) Set(phone, code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 覆盖式写入：重复发送时以最新验证码为准，避免日志与存储不一致。
	c.rdb.Set(ctx, c.key(phone), code, c.ttl)
}

func (c *RedisCodeCache) Verify(phone, code string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	k := c.key(phone)
	stored, err := c.rdb.Get(ctx, k).Result()
	if err != nil {
		// 不存在或已过期。
		return false
	}
	// 校验后立即删除，防止重放。
	c.rdb.Del(ctx, k)
	return stored == code
}

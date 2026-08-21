package svc

import (
	"sync"
	"time"
)

// codeEntry 是验证码缓存中的单条记录，保存验证码与过期时间。
type codeEntry struct {
	code      string
	expiresAt time.Time
}

// CodeCache 是本地内存验证码存储，用于联调阶段临时顶替短信/缓存服务。
// 注意：仅适用于单实例本地联调；多实例或生产环境应替换为 Redis 等共享存储。
type CodeCache struct {
	mu      sync.RWMutex
	entries map[string]codeEntry
	ttl     time.Duration
}

// NewCodeCache 创建验证码缓存，ttl 为验证码有效期。
func NewCodeCache(ttl time.Duration) *CodeCache {
	return &CodeCache{
		entries: make(map[string]codeEntry),
		ttl:     ttl,
	}
}

// TTL 返回验证码有效期。
func (c *CodeCache) TTL() time.Duration {
	return c.ttl
}

// Set 保存指定手机号的验证码。
func (c *CodeCache) Set(phone, code string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[phone] = codeEntry{code: code, expiresAt: time.Now().Add(c.ttl)}
}

// Verify 校验手机号验证码，匹配且未过期返回 true，并立即作废（防止重放）。
func (c *CodeCache) Verify(phone, code string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[phone]
	if !ok {
		return false
	}
	delete(c.entries, phone)
	return entry.code == code && time.Now().Before(entry.expiresAt)
}

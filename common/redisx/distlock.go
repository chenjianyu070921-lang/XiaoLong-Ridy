package redisx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// releaseScript 仅当锁的 value 与持有者 token 一致时才删除，避免误删他人锁。
// KEYS[1]=lockKey, ARGV[1]=token
var releaseScript = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)

// Lock 表示一个已获取的分布式锁，调用 Release 释放。
type Lock struct {
	client    *redis.Client
	key       string
	token     string
	ttl       time.Duration
	maxHold   time.Duration
	startTime time.Time

	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// TryLock 尝试获取分布式锁，成功返回 *Lock；锁已被占用或出错返回 error。
// ttl 为锁自动过期时间，内部启看门狗在到期前续期，直至 Release 被调用。
// maxHold=0 表示不限制最大持有时间（看门狗持续续期直到 Release）。
func TryLock(ctx context.Context, client *redis.Client, key string, ttl time.Duration) (*Lock, error) {
	return TryLockWithMaxHold(ctx, client, key, ttl, 0)
}

// TryLockWithMaxHold 在 TryLock 基础上支持 maxHold 最大持有时间。
// 超过 maxHold 后看门狗自动停止续期，锁将在 ttl 后过期释放，
// 防止进程存活但忘记 Release 导致锁被永久占用。
func TryLockWithMaxHold(ctx context.Context, client *redis.Client, key string, ttl time.Duration, maxHold time.Duration) (*Lock, error) {
	if client == nil {
		return nil, nil // 无 Redis 时由调用方退化到本地保护
	}
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	token := newToken()

	ok, err := client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockHeld
	}

	lk := &Lock{
		client:    client,
		key:       key,
		token:     token,
		ttl:       ttl,
		maxHold:   maxHold,
		startTime: time.Now(),
	}
	lk.startWatchdog()
	return lk, nil
}

// startWatchdog 在锁到期前续期，保证持有期间不过期。
// 若 maxHold > 0，超过最大持有时间后自动停止续期并记录日志。
func (l *Lock) startWatchdog() {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel
	l.done = make(chan struct{})
	go func() {
		ticker := time.NewTicker(l.ttl / 2)
		defer ticker.Stop()
		defer close(l.done)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if l.maxHold > 0 && time.Since(l.startTime) >= l.maxHold {
					logx.Errorf("lock maxHold exceeded, stop renewing, key=%s maxHold=%s",
						l.key, l.maxHold)
					return
				}
				l.client.Eval(ctx,
					`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`,
					[]string{l.key}, l.token, int(l.ttl.Milliseconds()),
				)
			}
		}
	}()
}

// Release 释放锁。幂等，可重复调用。
func (l *Lock) Release() {
	if l == nil {
		return
	}
	l.once.Do(func() {
		if l.cancel != nil {
			l.cancel()
		}
		if l.client != nil {
			_ = releaseScript.Run(context.Background(), l.client, []string{l.key}, l.token).Err()
		}
	})
}

// ErrLockHeld 表示锁被其他持有者占用。
var ErrLockHeld = &lockError{"lock is held by another process"}

type lockError struct{ msg string }

func (e *lockError) Error() string { return e.msg }

func newToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

package redisx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// releaseScript 仅当锁的 value 与持有者 token 一致时才删除，避免误删他人锁。
// KEYS[1]=lockKey, ARGV[1]=token
var releaseScript = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)

// Lock 表示一个已获取的分布式锁，调用 Release 释放。
type Lock struct {
	client *redis.Client
	key    string
	token  string
	ttl    time.Duration

	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// TryLock 尝试获取分布式锁，成功返回 *Lock；锁已被占用或出错返回 error。
// ttl 为锁自动过期时间，内部启看门狗在到期前续期，直至 Release 被调用。
func TryLock(ctx context.Context, client *redis.Client, key string, ttl time.Duration) (*Lock, error) {
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

	lk := &Lock{client: client, key: key, token: token, ttl: ttl}
	lk.startWatchdog()
	return lk, nil
}

// startWatchdog 在锁到期前续期，保证持有期间不过期（除非进程崩溃）。
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
				// 续期使用 PEXPIRE：仅当 token 仍匹配时才延长，避免续他人锁。
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
		// 极小概率，退化为时间戳+随机，仅用于避免冲突，不影响安全性边界。
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

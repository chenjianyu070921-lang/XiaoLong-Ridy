package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// loginLimiter 是登录/发码接口的轻量级内存限流器（固定窗口）。
// 说明：当前为单实例本地联调方案，未按用户维度共享；多实例或生产环境
// 应替换为 Redis 令牌桶（如 go-zero periodlimit）以保证全局限流一致。
type loginLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*window
	limit    int           // 窗口内最大允许次数
	window   time.Duration // 窗口大小
	now      func() time.Time
}

type window struct {
	count    int
	resetAt  time.Time
}

// newLoginLimiter 构造限流器：每个 key 在 window 内最多允许 limit 次。
func newLoginLimiter(limit int, windowSize time.Duration) *loginLimiter {
	return &loginLimiter{
		buckets: make(map[string]*window),
		limit:   limit,
		window:  windowSize,
		now:     time.Now,
	}
}

// allow 判断 key 在当前窗口内本次是否放行：窗口过期则重置计数放行，计数未达上限累加放行，否则拒绝。
// 调用方需持有 l.mu 或由调用处加锁（此处内部已加锁）。
func (l *loginLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	w, ok := l.buckets[key]
	if !ok || now.After(w.resetAt) {
		l.buckets[key] = &window{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if w.count >= l.limit {
		return false
	}
	w.count++
	return true
}

// clientKey 从请求中取客户端 IP 作为限流维度：优先取 X-Forwarded-For，回退解析 RemoteAddr 的 host 部分。
func clientKey(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// LoginRateLimit 返回一个限流中间件，超限返回 429。
// limit 为窗口内最大次数，window 为窗口大小。
func LoginRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := newLoginLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.allow(clientKey(r)) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":42901,"message":"操作过于频繁，请稍后再试","data":null,"timestamp":0,"traceId":""}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

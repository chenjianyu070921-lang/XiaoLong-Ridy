// Package middleware 提供司机端 HTTP 鉴权中间件。
package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	"XiaoLong-Ridy/common/jwtx"
)

// contextKey 是存放司机身份 claims 的 context key 类型，避免与其他包冲突。
type contextKey string

// ClaimsContextKey 是 claims 在 request context 中的键。
const ClaimsContextKey contextKey = "driverClaims"

// 司机账号状态（与 driversvc.DriverStatus 枚举值对齐：PENDING=1 NORMAL=2 FROZEN=3 CANCELLED=4）。
const (
	driverStatusFrozen    = 3 // 冻结
	driverStatusCancelled = 4 // 注销
)

// 封禁状态本地缓存 TTL：避免每次请求都查 driversvc，同时保证封禁后最多 30s 内全量生效。
// 生产环境应改为 Redis 黑名单或 driversvc 推送失效事件，做到秒级/即时生效。
const driverStatusCacheTTL = 30 * time.Second

// statusCache 缓存 driverID -> 最近一次查到的账号状态及过期时间（单机本地，降级方案）。
type statusCache struct {
	mu      sync.RWMutex
	entries map[uint64]statusEntry
}

type statusEntry struct {
	status   int
	expireAt time.Time
}

func newStatusCache() *statusCache {
	return &statusCache{entries: make(map[uint64]statusEntry)}
}

// get 返回缓存的状态；过期或不存在返回 (0, false)。
func (c *statusCache) get(id uint64) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[id]
	if !ok || time.Now().After(e.expireAt) {
		return 0, false
	}
	return e.status, true
}

// set 写入状态并重置 TTL。
func (c *statusCache) set(id uint64, status int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[id] = statusEntry{status: status, expireAt: time.Now().Add(driverStatusCacheTTL)}
}

// RequireAuth 返回一个 HTTP 中间件：校验 Bearer JWT，失败返回 401。
// 校验通过后，将解析出的司机身份 claims 注入 request context。
func RequireAuth(svcCtx *svc.ServiceContext) func(http.Handler) http.Handler {
	// 本地封禁状态缓存（单实例降级方案，见 statusCache 注释）。
	cache := newStatusCache()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r.Header.Get("Authorization"))
			if token == "" {
				writeUnauthorized(w, "缺少登录凭证")
				return
			}
			claims, err := jwtx.ParseAccountToken(token, svcCtx.SigningKey)
			if err == jwtx.ErrTokenExpired {
				writeUnauthorized(w, "登录凭证已过期")
				return
			}
			if err != nil || claims.AccountType != "driver" {
				writeUnauthorized(w, "登录凭证无效")
				return
			}
			// JWT 内已为封禁/注销态直接拒绝（封禁前签发的旧 token 也能在此拦截）。
			if claims.AccountStatus == driverStatusFrozen || claims.AccountStatus == driverStatusCancelled {
				writeForbidden(w, "账号已被冻结或注销")
				return
			}
			// 实时校验账号最新状态：命中本地缓存的封禁态直接拒绝；
			// 否则查 driversvc 获取最新状态（带超时，失败降级为信任 JWT）。
			if cached, ok := cache.get(claims.AccountID); ok && isBlockedStatus(cached) {
				writeForbidden(w, "账号已被冻结或注销")
				return
			}
			if svcCtx.DriverClient != nil {
				if st, ok := cache.get(claims.AccountID); !ok {
					dCtx, dCancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
					defer dCancel()
					if d, dErr := svcCtx.DriverClient.GetDriver(dCtx, &driversproto.GetDriverRequest{Id: int64(claims.AccountID)}); dErr == nil {
						st := int(d.GetDriver().GetStatus())
						cache.set(claims.AccountID, st)
						if isBlockedStatus(st) {
							writeForbidden(w, "账号已被冻结或注销")
							return
						}
					}
				} else if isBlockedStatus(st) {
					writeForbidden(w, "账号已被冻结或注销")
					return
				}
			}
			// 将 claims 注入 context，供下游 handler 读取司机身份。
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isBlockedStatus 判断司机状态是否为封禁/注销（不可继续服务）。
func isBlockedStatus(status int) bool {
	return status == driverStatusFrozen || status == driverStatusCancelled
}

// extractBearer 从 Authorization 头中提取 Bearer token。
func extractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

// writeUnauthorized 写出统一的 401 未授权响应。
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"code":40102,"message":"` + message + `","data":null,"timestamp":0,"traceId":""}`))
}

// writeForbidden 写出统一的 403 禁止响应（账号封禁/注销场景）。
func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"code":40301,"message":"` + message + `","data":null,"timestamp":0,"traceId":""}`))
}

// ClaimsFromContext 从 request context 中取出司机身份 claims，未登录返回 nil。
func ClaimsFromContext(ctx context.Context) *jwtx.AccountClaims {
	claims, ok := ctx.Value(ClaimsContextKey).(*jwtx.AccountClaims)
	if !ok {
		return nil
	}
	return claims
}

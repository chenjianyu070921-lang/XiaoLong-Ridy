package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
)

const (
	InternalServiceTokenHeader = "X-Internal-Service-Token"
	TraceIDHeader              = "X-Trace-Id"
	internalCaller             = "internal_admin"
)

const (
	internalCallContextKey   contextKey = "internalCall"
	internalTraceContextKey  contextKey = "internalTraceId"
	internalCallerContextKey contextKey = "caller"
)

type internalAuthMiddleware struct {
	serviceToken string
	allowed      map[string]struct{}
	limiter      *loginLimiter
}

type traceResponseWriter struct {
	http.ResponseWriter
	traceID string
}

func (w *traceResponseWriter) ResponseTraceID() string {
	return w.traceID
}

func InternalServiceAuth(svcCtx *svc.ServiceContext) func(http.Handler) http.Handler {
	cfg := svc.InternalAuthConfig{}
	if svcCtx != nil {
		cfg = svcCtx.InternalAuth
	}
	token := strings.TrimSpace(cfg.ServiceToken)
	allowed := make(map[string]struct{}, len(cfg.AllowedRoutes))
	for _, route := range cfg.AllowedRoutes {
		allowed[routeKey(route.Method, route.Path)] = struct{}{}
	}
	var limiter *loginLimiter
	if token != "" {
		limit := cfg.RateLimit.Limit
		windowSeconds := cfg.RateLimit.WindowSeconds
		if limit <= 0 {
			limit = 60
		}
		if windowSeconds <= 0 {
			windowSeconds = 60
		}
		limiter = newLoginLimiter(limit, time.Duration(windowSeconds)*time.Second)
	}
	m := &internalAuthMiddleware{
		serviceToken: token,
		allowed:      allowed,
		limiter:      limiter,
	}
	return m.handle
}

func (m *internalAuthMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presented := r.Header.Get(InternalServiceTokenHeader)
		if presented == "" {
			next.ServeHTTP(w, r)
			return
		}
		if m.serviceToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !constantTimeTokenEqual(presented, m.serviceToken) {
			writeForbidden(w, "internal service token invalid")
			return
		}
		if _, ok := m.allowed[routeKey(r.Method, r.URL.Path)]; !ok {
			writeForbidden(w, "internal service route forbidden")
			return
		}
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = r.Header.Get("X-Trace-ID")
		}
		traceID = strings.TrimSpace(traceID)
		if traceID == "" {
			writeBadRequest(w, "internal traceId required")
			return
		}
		if m.limiter != nil && !m.limiter.allow(clientKey(r)+" "+r.Method+" "+r.URL.Path) {
			writeTooManyRequests(w, "internal service rate limit exceeded", traceID)
			return
		}
		ctx := context.WithValue(r.Context(), internalCallContextKey, true)
		ctx = context.WithValue(ctx, internalTraceContextKey, traceID)
		ctx = context.WithValue(ctx, internalCallerContextKey, internalCaller)
		log.Printf(`caller="%s" method="%s" path="%s" traceId="%s"`, internalCaller, r.Method, r.URL.Path, traceID)
		next.ServeHTTP(&traceResponseWriter{ResponseWriter: w, traceID: traceID}, r.WithContext(ctx))
	})
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func constantTimeTokenEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}

func IsInternalCall(ctx context.Context) bool {
	v, _ := ctx.Value(internalCallContextKey).(bool)
	return v
}

func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(internalTraceContextKey).(string)
	return v
}

func CallerFromContext(ctx context.Context) string {
	v, _ := ctx.Value(internalCallerContextKey).(string)
	return v
}

func writeBadRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"code":50000,"message":"` + message + `","data":null,"timestamp":0,"traceId":""}`))
}

func writeTooManyRequests(w http.ResponseWriter, message, traceID string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"code":42901,"message":"` + message + `","data":null,"timestamp":0,"traceId":"` + traceID + `"}`))
}

// Package middleware 提供司机端 HTTP 鉴权中间件。
package middleware

import (
	"context"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/common/jwtx"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// contextKey 是存放司机身份 claims 的 context key 类型，避免与其他包冲突。
type contextKey string

// ClaimsContextKey 是 claims 在 request context 中的键。
const ClaimsContextKey contextKey = "driverClaims"

// RequireAuth 返回一个 HTTP 中间件：校验 Bearer JWT，失败返回 401。
// 校验通过后，将解析出的司机身份 claims 注入 request context。
func RequireAuth(svcCtx *svc.ServiceContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if IsInternalCall(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
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
			if claims.AccountStatus != int(driversproto.DriverStatus_DRIVER_STATUS_NORMAL) &&
				(claims.AccountStatus != int(driversproto.DriverStatus_DRIVER_STATUS_PENDING) || !allowPendingDriverPath(r.URL.Path)) {
				writeForbidden(w, "账号未通过审核或已被冻结/注销")
				return
			}
			// 将 claims 注入 context，供下游 handler 读取司机身份。
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func allowPendingDriverPath(path string) bool {
	switch path {
	case "/api/driver/v1/drivers/update",
		"/api/driver/v1/drivers/get",
		"/api/driver/v1/drivers/certification/upload",
		"/api/driver/v1/drivers/certification",
		"/api/driver/v1/vehicles",
		"/api/driver/v1/vehicles/get",
		"/api/driver/v1/vehicles/update",
		"/api/driver/v1/income/summary",
		"/api/driver/v1/income/today",
		"/api/driver/v1/income/week",
		"/api/driver/v1/income/bills":
		return true
	default:
		return false
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"code":40301,"message":"` + message + `","data":null,"timestamp":0,"traceId":""}`))
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

// ClaimsFromContext 从 request context 中取出司机身份 claims，未登录返回 nil。
func ClaimsFromContext(ctx context.Context) *jwtx.AccountClaims {
	claims, ok := ctx.Value(ClaimsContextKey).(*jwtx.AccountClaims)
	if !ok {
		return nil
	}
	return claims
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/chat/internal/svc"
	"XiaoLong-Ridy/common/jwtx"
)

type contextKey string

// ClaimsContextKey 是 claims 在 request context 中的键。
const ClaimsContextKey contextKey = "chatClaims"

// RequireAuth 校验 Bearer JWT（尝试多个签名密钥，兼容司机端与乘客端），
// 校验通过后把 claims 注入 context。仅允许 driver / passenger 两类账号。
func RequireAuth(svcCtx *svc.ServiceContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r.Header.Get("Authorization"))
			if token == "" {
				writeUnauthorized(w, "缺少登录凭证")
				return
			}
			var claims *jwtx.AccountClaims
			for _, key := range svcCtx.SigningKeys {
				if c, err := jwtx.ParseAccountToken(token, key); err == nil {
					claims = c
					break
				}
			}
			if claims == nil {
				writeUnauthorized(w, "登录凭证无效")
				return
			}
			if claims.AccountType != "driver" && claims.AccountType != "passenger" {
				writeUnauthorized(w, "登录凭证类型不支持")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"code":40102,"message":"` + message + `","data":null,"timestamp":0,"traceId":""}`))
}

// ClaimsFromContext 取出当前请求的身份 claims，未登录返回 nil。
func ClaimsFromContext(ctx context.Context) *jwtx.AccountClaims {
	c, ok := ctx.Value(ClaimsContextKey).(*jwtx.AccountClaims)
	if !ok {
		return nil
	}
	return c
}

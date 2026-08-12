package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// SendSMSCodeHandler 处理 POST /api/passenger/v1/auth/send-sms-code。
func SendSMSCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSMSCodeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAuthLogic(r.Context(), svcCtx).SendSMSCode(&req)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// LoginBySMSHandler 处理 POST /api/passenger/v1/auth/login-by-sms。
func LoginBySMSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginBySMSRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAuthLogic(r.Context(), svcCtx).LoginBySMS(&req)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// RefreshTokenHandler 处理 POST /api/passenger/v1/auth/refresh-token。
func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAuthLogic(r.Context(), svcCtx).RefreshToken(&req)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// LogoutHandler 处理 POST /api/passenger/v1/auth/logout。
func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 文档要求登出接口从 Authorization: Bearer {JWT} 中读取当前登录态。
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			writeError(w, http.StatusUnauthorized, 40102, "Token无效")
			return
		}
		resp, err := logic.NewAuthLogic(r.Context(), svcCtx).Logout(token)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, 50000, "仅支持POST请求")
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, 50000, "请求体不是合法JSON")
		return false
	}
	return true
}

func writeAuthError(w http.ResponseWriter, err error) {
	// 将 usersvc 的公开业务错误映射为接口文档约定的 HTTP 状态码与业务错误码。
	switch {
	case errors.Is(err, logic.ErrUserClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50001, "下游服务不可用")
	case errors.Is(err, userproto.ErrInvalidPhone):
		writeError(w, http.StatusBadRequest, 50000, "手机号格式不合法")
	case errors.Is(err, userproto.ErrInvalidSMSCode):
		writeError(w, http.StatusBadRequest, 41001, "验证码错误")
	case errors.Is(err, userproto.ErrSMSCodeExpired):
		writeError(w, http.StatusBadRequest, 41002, "验证码已过期")
	case errors.Is(err, userproto.ErrSMSCodeSendTooFrequent):
		writeError(w, http.StatusTooManyRequests, 41003, "验证码发送过于频繁")
	case errors.Is(err, userproto.ErrAccountFrozen):
		writeError(w, http.StatusForbidden, 40301, "账号已被封禁")
	case errors.Is(err, userproto.ErrTokenExpired):
		writeError(w, http.StatusUnauthorized, 40101, "Token已过期")
	case errors.Is(err, userproto.ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, 40102, "Token无效")
	default:
		writeError(w, http.StatusInternalServerError, 50000, "服务器内部错误")
	}
}

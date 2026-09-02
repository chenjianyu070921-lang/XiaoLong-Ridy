package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendSMSCodeHandler POST /api/driver/v1/auth/send-sms-code
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

// LoginByPasswordHandler POST /api/driver/v1/auth/login-by-password
func LoginByPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginByPasswordRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAuthLogic(r.Context(), svcCtx).LoginByPassword(&req)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// LoginBySMSHandler POST /api/driver/v1/auth/login-by-sms
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

// writeAuthError maps auth failures to HTTP responses.
func writeAuthError(w http.ResponseWriter, err error) {
	switch err {
	case logic.ErrDriverAuthFailed:
		writeError(w, http.StatusUnauthorized, 40102, "账号或密码错误")
	case logic.ErrDriverFrozen:
		writeError(w, http.StatusForbidden, 40301, "账号未审核通过或已被冻结/注销")
	case logic.ErrCodeInvalid:
		writeError(w, http.StatusBadRequest, 41001, "验证码错误")
	case logic.ErrDriverClientNotConfigured:
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	default:
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded, codes.Unimplemented:
				writeError(w, http.StatusBadGateway, 50001, "下游服务不可用或超时")
			case codes.PermissionDenied:
				writeError(w, http.StatusForbidden, 40301, "账号未审核通过或已被冻结/注销")
			default:
				writeError(w, http.StatusUnauthorized, 40102, "账号或密码错误")
			}
			return
		}
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	}
}

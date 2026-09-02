package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"

	"google.golang.org/grpc/status"
)

// SendSMSCodeHandler POST /api/driver/v1/auth/send-sms-code
// 发送登录短信验证码（联调阶段验证码输出到日志顶替短信通道）。
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
// 手机号 + 密码登录，成功返回 JWT。
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
// 手机号 + 验证码登录，成功返回 JWT。
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

// writeAuthError 将认证相关错误映射为统一错误响应。
func writeAuthError(w http.ResponseWriter, err error) {
	switch err {
	case logic.ErrDriverAuthFailed:
		writeError(w, http.StatusUnauthorized, 40102, "账号或密码错误")
	case logic.ErrDriverFrozen:
		writeError(w, http.StatusForbidden, 40301, "账号未审核通过或已被冻结/注销")
	case logic.ErrCodeInvalid:
		writeError(w, http.StatusBadRequest, 41001, "验证码错误")
	case logic.ErrCodeSendFailed:
		// 验证码生成或缓存不可用属于服务端依赖故障，返回 503 便于前端和监控正确识别。
		writeError(w, http.StatusServiceUnavailable, 50004, "验证码服务暂不可用")
	case logic.ErrDriverClientNotConfigured:
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	default:
		// gRPC status 透传（如 driversvc 返回的 driver not found）。
		if _, ok := status.FromError(err); ok {
			writeError(w, http.StatusUnauthorized, 40102, "账号或密码错误")
			return
		}
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	}
}

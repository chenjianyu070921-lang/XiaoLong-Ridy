package handler

import (
	"errors"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/status"
)

// bearerToken 从 Authorization 头中解析 Bearer Token。
func bearerToken(r *http.Request) string {
	return strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
}

// writeBusinessError 将业务层错误转换为统一 HTTP 响应。
func writeBusinessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, codeInvalidToken, "Token无效")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, codeForbidden, "无权访问该资源")
	case errors.Is(err, logic.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "请求参数不合法")
	case errors.Is(err, logic.ErrOrderClientNotConfigured),
		errors.Is(err, logic.ErrPriceClientNotConfigured),
		errors.Is(err, logic.ErrUserClientNotConfigured):
		writeError(w, http.StatusBadGateway, codeDownstreamUnavailable, "下游服务不可用")
	case matchesBusinessError(err, userproto.ErrInvalidAddressPhone):
		writeError(w, http.StatusBadRequest, codeInvalidAddressPhone, "异常_电话格式错误")
	case matchesBusinessError(err, userproto.ErrInvalidLongitudeLatitude):
		writeError(w, http.StatusBadRequest, codeInvalidLongitudeLatitude, "异常_经纬度错误")
	case matchesBusinessError(err, userproto.ErrAddressNotFound):
		writeError(w, http.StatusNotFound, codeAddressNotFound, "地址不存在")
	case matchesBusinessError(err, userproto.ErrUserNotFound):
		writeError(w, http.StatusNotFound, codeUserNotFound, "用户不存在")
	default:
		writeError(w, http.StatusInternalServerError, codeInternalError, "服务器内部错误")
	}
}

// matchesBusinessError 同时兼容本地直调的 sentinel error 和真实 gRPC status message。
func matchesBusinessError(err error, target error) bool {
	if errors.Is(err, target) {
		return true
	}
	return status.Convert(err).Message() == target.Error()
}

package handler

import (
	"errors"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// bearerToken 从 Authorization 头中解析 Bearer Token。
func bearerToken(r *http.Request) string {
	return strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
}

// writeBusinessError 将业务层错误转换为统一 HTTP 响应。
func writeBusinessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, 40102, "Token无效")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, 40303, "无权访问该资源")
	case errors.Is(err, logic.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, 50000, "请求参数不合法")
	case errors.Is(err, logic.ErrOrderClientNotConfigured),
		errors.Is(err, logic.ErrPriceClientNotConfigured),
		errors.Is(err, logic.ErrUserClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50001, "下游服务不可用")
	case errors.Is(err, userproto.ErrInvalidAddressPhone):
		writeError(w, http.StatusBadRequest, 41001, "异常_电话格式错误")
	case errors.Is(err, userproto.ErrInvalidLongitudeLatitude):
		writeError(w, http.StatusBadRequest, 41004, "异常_经纬度为0")
	case errors.Is(err, userproto.ErrAddressNotFound):
		writeError(w, http.StatusNotFound, 40401, "地址不存在")
	default:
		writeError(w, http.StatusInternalServerError, 50000, "服务器内部错误")
	}
}

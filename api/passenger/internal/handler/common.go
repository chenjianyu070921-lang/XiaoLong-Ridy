package handler

import (
	"errors"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/codes"
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
		writeError(w, http.StatusUnauthorized, codeInvalidToken, "Token invalid")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, codeForbidden, "forbidden")
	case errors.Is(err, logic.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid request")
	case errors.Is(err, logic.ErrOrderNotPayable):
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "order not payable")
	case errors.Is(err, logic.ErrOrderClientNotConfigured),
		errors.Is(err, logic.ErrPriceClientNotConfigured),
		errors.Is(err, logic.ErrPayClientNotConfigured),
		errors.Is(err, logic.ErrDispatchClientNotConfigured),
		errors.Is(err, logic.ErrUserClientNotConfigured),
		isDownstreamGRPCError(err):
		writeError(w, http.StatusBadGateway, codeDownstreamUnavailable, "downstream unavailable")
	case matchesBusinessError(err, userproto.ErrInvalidAddressPhone):
		writeError(w, http.StatusBadRequest, codeInvalidAddressPhone, "invalid address phone")
	case matchesBusinessError(err, userproto.ErrInvalidLongitudeLatitude):
		writeError(w, http.StatusBadRequest, codeInvalidLongitudeLatitude, "invalid longitude or latitude")
	case matchesBusinessError(err, userproto.ErrAddressNotFound):
		writeError(w, http.StatusNotFound, codeAddressNotFound, "address not found")
	case matchesBusinessError(err, userproto.ErrUserNotFound):
		writeError(w, http.StatusNotFound, codeUserNotFound, "user not found")
	case matchesBusinessError(err, userproto.ErrCouponNotFound):
		writeError(w, http.StatusNotFound, codeCouponNotFound, "coupon not found")
	case matchesBusinessError(err, userproto.ErrUserCouponNotFound):
		writeError(w, http.StatusNotFound, codeCouponNotFound, "user coupon not found")
	case matchesBusinessError(err, userproto.ErrCouponUnavailable):
		writeError(w, http.StatusBadRequest, codeCouponUnavailable, "coupon unavailable")
	case matchesBusinessError(err, userproto.ErrCouponReceiveLimit):
		writeError(w, http.StatusBadRequest, codeCouponReceiveLimit, "coupon receive limit exceeded")
	case matchesBusinessErrorMessage(err, "invalid real name info"):
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid real name info")
	case matchesBusinessErrorMessage(err, "real name verification failed"):
		writeError(w, http.StatusBadRequest, codeRealNameVerifyFailed, "real name verification failed")
	case errors.Is(err, logic.ErrReviewAlreadyExists):
		writeError(w, http.StatusBadRequest, codeReviewAlreadyExists, "review already exists")
	case errors.Is(err, logic.ErrReviewRepositoryNotConfigured):
		writeError(w, http.StatusBadGateway, codeDownstreamUnavailable, "review repository not configured")
	case errors.Is(err, logic.ErrQiniuNotConfigured):
		writeError(w, http.StatusBadGateway, codeDownstreamUnavailable, "qiniu storage not configured")
	default:
		writeError(w, http.StatusInternalServerError, codeInternalError, "internal error")
	}
}

// isDownstreamGRPCError 判断错误是否来自下游 RPC 服务不可达或超时。
func isDownstreamGRPCError(err error) bool {
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

// matchesBusinessError 同时兼容本地直调的 sentinel error 和真实 gRPC status message。
func matchesBusinessError(err error, target error) bool {
	if errors.Is(err, target) {
		return true
	}
	return status.Convert(err).Message() == target.Error()
}

// matchesBusinessErrorMessage 兼容下游 RPC 将普通 Go error 包装为 gRPC status 的场景。
func matchesBusinessErrorMessage(err error, target string) bool {
	if err == nil {
		return false
	}
	if err.Error() == target {
		return true
	}
	return status.Convert(err).Message() == target
}

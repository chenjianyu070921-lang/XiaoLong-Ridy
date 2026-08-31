package handler

import (
	"errors"
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/logic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	codeDriverAlreadyExists = 50010
	codeDriverNotFound      = 50011
)

func writeParamError(w http.ResponseWriter, err error) {
	if errors.Is(err, logic.ErrDriverClientNotConfigured) {
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
		return
	}
	if errors.Is(err, logic.ErrOrderClientNotConfigured) {
		writeError(w, http.StatusBadGateway, 50002, "下游 ordersvc 不可用")
		return
	}
	if errors.Is(err, logic.ErrReviewStorageNotConfigured) || errors.Is(err, logic.ErrTrajectoryStorageNotConfigured) {
		writeError(w, http.StatusNotImplemented, 50003, err.Error())
		return
	}
	if errors.Is(err, logic.ErrInvalidParam) {
		writeError(w, http.StatusBadRequest, 50000, err.Error())
		return
	}
	if errors.Is(err, logic.ErrForbiddenDriverResource) {
		writeError(w, http.StatusForbidden, 40301, "无权访问该司机资源")
		return
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable:
			writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
			return
		case codes.NotFound:
			writeError(w, http.StatusNotFound, codeDriverNotFound, st.Message())
			return
		case codes.AlreadyExists:
			writeError(w, http.StatusConflict, codeDriverAlreadyExists, "手机号或驾驶证号已存在")
			return
		case codes.PermissionDenied:
			writeError(w, http.StatusForbidden, 40301, "无权访问该司机资源")
			return
		case codes.Unknown:
			if strings.Contains(st.Message(), "already exists") {
				writeError(w, http.StatusConflict, codeDriverAlreadyExists, "手机号或驾驶证号已存在")
				return
			}
		}
		writeError(w, http.StatusBadRequest, 50000, st.Message())
		return
	}

	writeError(w, http.StatusBadRequest, 50000, err.Error())
}

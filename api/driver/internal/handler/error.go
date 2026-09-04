package handler

import (
	"errors"
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	codeDriverAlreadyExists  = 50010
	codeDriverNotFound       = 50011
	codeInternalServerError  = 50006
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
	if errors.Is(err, logic.ErrPayClientNotConfigured) {
		writeError(w, http.StatusBadGateway, 50003, "下游 paysvc 不可用")
		return
	}
	if errors.Is(err, logic.ErrPriceClientNotConfigured) {
		writeError(w, http.StatusBadGateway, 50005, "下游 pricesvc 不可用")
		return
	}
	if errors.Is(err, logic.ErrQiniuNotConfigured) {
		writeError(w, http.StatusBadGateway, 50004, "qiniu storage not configured")
		return
	}
	if errors.Is(err, logic.ErrReviewRepositoryNotConfigured) {
		writeError(w, http.StatusServiceUnavailable, 50001, "评价存储不可用")
		return
	}
	if errors.Is(err, logic.ErrTrajectoryRepositoryNotConfigured) {
		writeError(w, http.StatusServiceUnavailable, 50001, "轨迹存储不可用")
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
		case codes.Unavailable, codes.DeadlineExceeded, codes.Unimplemented:
			// 下游服务不可用/超时统一映射为 502，避免被误判为 400 参数错误。
			writeError(w, http.StatusBadGateway, 50001, "下游服务不可用或超时")
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
			// 下游以裸 errors.New 返回的错误经 gRPC 包成 Unknown，其消息多为内部实现细节，
			// 不得透传给司机端，统一返回通用错误避免信息泄漏。
			writeError(w, http.StatusInternalServerError, codeInternalServerError, "服务暂时不可用，请稍后重试")
			return
		default:
			// 其他显式 gRPC 错误码（InvalidArgument/Unauthenticated/FailedPrecondition 等）：
			// 消息由 driversvc 显式声明为可展示给司机的业务提示，安全透传。
			writeError(w, http.StatusBadRequest, 50000, st.Message())
			return
		}
	}

	// 既非已识别的 sentinel，也非 gRPC status 错误：按内部错误收口，不透传原始消息。
	writeError(w, http.StatusInternalServerError, codeInternalServerError, "服务暂时不可用，请稍后重试")
}

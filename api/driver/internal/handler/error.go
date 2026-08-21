// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"errors"   // 用于错误类型判断
	"net/http" // 提供 HTTP 状态码
	"strings"

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层错误定义
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	codeDriverAlreadyExists = 40901
	codeDriverNotFound      = 40401
)

// writeParamError 将 logic 层返回的业务/校验错误映射为统一的 HTTP 响应。
func writeParamError(w http.ResponseWriter, err error) {
	// 按错误类型分支处理。
	switch {
	// 下游 driversvc 客户端不可用：返回 502 + 业务码 50001。
	case errors.Is(err, logic.ErrDriverClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	// 下游 ordersvc 客户端不可用：返回 502 + 业务码 50002。
	case errors.Is(err, logic.ErrOrderClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50002, "下游 ordersvc 不可用")
	case errors.Is(err, logic.ErrDispatchClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50004, "下游 dispatchsvc 不可用")
	// 通用参数错误：返回 400 + 业务码 50000，并透传错误文案。
	case errors.Is(err, logic.ErrInvalidParam):
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	// 默认分支：logic 直接返回的中文校验错误，统一归为 400 参数错误。
	default:
		writeDriverRPCError(w, err)
	}
}

func writeDriverRPCError(w http.ResponseWriter, err error) {
	grpcStatus := status.Convert(err)
	message := grpcStatus.Message()
	if message == "driver already exists" {
		writeError(w, http.StatusConflict, codeDriverAlreadyExists, "手机号或驾驶证号已存在")
		return
	}

	switch grpcStatus.Code() {
	case codes.AlreadyExists:
		writeError(w, http.StatusConflict, codeDriverAlreadyExists, message)
	case codes.NotFound:
		writeError(w, http.StatusNotFound, codeDriverNotFound, message)
	case codes.Unauthenticated:
		writeError(w, http.StatusUnauthorized, 40102, message)
	case codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40301, message)
	case codes.Unavailable, codes.DeadlineExceeded:
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	case codes.Internal, codes.DataLoss:
		writeError(w, http.StatusInternalServerError, 50003, "司机服务处理失败")
	default:
		writeError(w, http.StatusBadRequest, 50000, strings.TrimSpace(message))
	}
}

// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"errors"   // 用于错误类型判断
	"net/http" // 提供 HTTP 状态码

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层错误定义
)

// writeParamError 将 logic 层返回的业务/校验错误映射为统一的 HTTP 响应。
func writeParamError(w http.ResponseWriter, err error) {
	// 按错误类型分支处理。
	switch {
	// 下游 driversvc 客户端不可用：返回 502 + 业务码 50001。
	case errors.Is(err, logic.ErrDriverClientNotConfigured):
		writeError(w, http.StatusBadGateway, 50001, "下游 driversvc 不可用")
	// 通用参数错误：返回 400 + 业务码 50000，并透传错误文案。
	case errors.Is(err, logic.ErrInvalidParam):
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	// 默认分支：logic 直接返回的中文校验错误，统一归为 400 参数错误。
	default:
		writeError(w, http.StatusBadRequest, 50000, err.Error())
	}
}

package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"XiaoLong-Ridy/api/passenger/internal/types"
)

// writeSuccess 返回业务成功响应，data 会被序列化到统一响应体的 data 字段。
func writeSuccess(w http.ResponseWriter, data any) {
	writeResponse(w, http.StatusOK, codeSuccess, "success", data)
}

// writeError 返回业务失败响应；status 是 HTTP 状态码，code 是接口文档约定的业务错误码。
func writeError(w http.ResponseWriter, status, code int, message string) {
	writeResponse(w, status, code, message, nil)
}

// writeResponse 统一写出 JSON 响应体。所有 handler 都必须经此出口，
// 保证前端始终拿到固定的 {code, message, data, timestamp, traceId} 结构。
func writeResponse(w http.ResponseWriter, status, code int, message string, data any) {
	// Content-Type 必须在 WriteHeader 之前设置：一旦 WriteHeader 调用，响应头即发出，再设置会被忽略。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// 序列化/写回失败时状态码已发出，无法再改为 5xx，此处静默忽略是刻意为之。
	_ = json.NewEncoder(w).Encode(types.Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		TraceID:   traceID(),
	})
}

// traceID 生成本次请求的追踪 ID（8 字节随机数），便于前后端串联同一次请求的日志。
// 极少数情况下系统熵源不可用，降级为固定值 trace_local，保证响应体不缺失该字段。
func traceID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "trace_local"
	}
	return "trace_" + hex.EncodeToString(value[:])
}

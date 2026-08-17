// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"crypto/rand"  // 用于生成随机 traceID
	"encoding/hex" // 将随机字节编码为十六进制字符串
	"encoding/json" // 提供 JSON 编解码能力
	"net/http"      // HTTP 基础类型与状态码
	"strconv"       // 将查询参数字符串解析为整型
	"time"          // 生成响应时间戳

	"XiaoLong-Ridy/api/driver/internal/types" // 统一响应结构体定义
)

// writeSuccess 以统一格式输出成功响应，HTTP 200 + code 0。
func writeSuccess(w http.ResponseWriter, data any) {
	// 调用底层写入函数，成功时 code=0、message="success"。
	writeResponse(w, http.StatusOK, 0, "success", data)
}

// writeError 以统一格式输出错误响应，data 为 nil。
func writeError(w http.ResponseWriter, status, code int, message string) {
	// 调用底层写入函数，携带 HTTP 状态码、业务码与错误信息。
	writeResponse(w, status, code, message, nil)
}

// writeResponse 是统一的响应写入函数，负责设置头、状态码与 JSON 编码。
func writeResponse(w http.ResponseWriter, status, code int, message string, data any) {
	// 设置响应内容类型为 JSON（含 UTF-8）。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 写入 HTTP 状态码。
	w.WriteHeader(status)
	// 将统一响应结构编码为 JSON 并写入响应体（忽略编码错误）。
	_ = json.NewEncoder(w).Encode(types.Response{
		Code:      code,             // 业务码
		Message:   message,          // 提示信息
		Data:      data,             // 业务数据（可为 nil）
		Timestamp: time.Now().Unix(), // 当前秒级时间戳
		TraceID:   traceID(),         // 本次请求追踪 ID
	})
}

// traceID 生成一个随机的 traceID 用于请求链路追踪。
func traceID() string {
	// 8 字节随机缓冲。
	var value [8]byte
	// 读取随机数填充缓冲；失败则回退到固定值。
	if _, err := rand.Read(value[:]); err != nil {
		return "trace_local"
	}
	// 将随机字节编码为十六进制并拼接前缀。
	return "trace_" + hex.EncodeToString(value[:])
}

// decodeJSON 校验请求方法为 POST 并解析 JSON 请求体到 target。解析失败已直接写错误响应。
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	// 仅允许 POST 方法。
	if r.Method != http.MethodPost {
		// 方法不支持，返回 405。
		writeError(w, http.StatusMethodNotAllowed, 50000, "仅支持POST请求")
		return false
	}
	// 解析 JSON 请求体到目标结构。
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		// 解析失败，返回 400 与错误信息。
		writeError(w, http.StatusBadRequest, 50000, "请求体不是合法JSON")
		return false
	}
	// 解析成功。
	return true
}

// decodeQueryID 从 URL 查询参数中读取整型 id，缺失或非法时返回 (0, false)。
func decodeQueryID(r *http.Request, key string) (int64, bool) {
	// 读取指定 key 的查询参数原始字符串。
	raw := r.URL.Query().Get(key)
	// 参数缺失直接返回失败。
	if raw == "" {
		return 0, false
	}
	// 将字符串解析为 64 位整数。
	id, err := strconv.ParseInt(raw, 10, 64)
	// 解析失败或值非正，返回失败。
	if err != nil || id <= 0 {
		return 0, false
	}
	// 解析成功且合法，返回 id 与 true。
	return id, true
}

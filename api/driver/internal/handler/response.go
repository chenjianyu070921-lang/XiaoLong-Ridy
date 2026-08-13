package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"XiaoLong-Ridy/api/driver/internal/types"
)

func writeSuccess(w http.ResponseWriter, data any) {
	writeResponse(w, http.StatusOK, 0, "success", data)
}

func writeError(w http.ResponseWriter, status, code int, message string) {
	writeResponse(w, status, code, message, nil)
}

func writeResponse(w http.ResponseWriter, status, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		TraceID:   traceID(),
	})
}

func traceID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "trace_local"
	}
	return "trace_" + hex.EncodeToString(value[:])
}

// decodeJSON 校验 POST + 解析 JSON 请求体。
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, 50000, "仅支持POST请求")
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, 50000, "请求体不是合法JSON")
		return false
	}
	return true
}

// decodeQueryID 从 URL 查询参数读取整型 id，缺失或非法时返回 (0, false)。
func decodeQueryID(r *http.Request, key string) (int64, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
